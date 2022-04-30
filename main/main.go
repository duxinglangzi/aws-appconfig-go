package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"log"
	"time"
)

var (
	versionsMap            = make(map[string]int32)
	applicationPipelineMap = make(map[string]string)
)

const (
	BaseApplicationName        = "duxinglangzi_application"
	BaseApplicationProfileName = "pipeline_config"
	LargeCycle                 = 300
	SmallCycle                 = 10
	Cycle                      = 12
)

func main() {
	
	log.Println("启动小程序,开始检查app config配置内容...")
	
	// defaultConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"),
	// 	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("aaaa", "bb", "")))
	
	defaultConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}
	
	// 读取所有的application
	appConfigClient := appconfig.NewFromConfig(defaultConfig)
	
	// codepipelineClient := codepipeline.NewFromConfig(defaultConfig)
	
	i := 1
	for true {
		applications, _ := appConfigClient.ListApplications(context.TODO(), nil, func(*appconfig.Options) {})
		// filterBaseApplication(applications, appConfigClient) // 加载最新的项目配置
		for _, item := range applications.Items {
			// execCheckApplication(&item, appConfigClient, codepipelineClient)
			log.Println("处理应用名称:", *item.Name, " , id:", *item.Id)
		}
		log.Println("本轮检查应用程序数量:", len(applications.Items))
		doSleep()
		i++
		if i > 5 {
			log.Println("关闭了.....")
			break
		}
	}
}

func doSleep() {
	// init the shanghai loc
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if time.Now().In(loc).Hour() < Cycle {
		time.Sleep(time.Second * time.Duration(LargeCycle))
	} else {
		time.Sleep(time.Second * time.Duration(SmallCycle))
	}
}

func execCheckApplication(application *types.Application, client *appconfig.Client, pipelineClient *codepipeline.Client) {
	// 查询应用下面的所有配置文件
	profiles, _ := client.ListConfigurationProfiles(context.TODO(), &appconfig.ListConfigurationProfilesInput{ApplicationId: application.Id})
	if profiles == nil {
		return
	}
	for _, item := range profiles.Items {
		configurationVersions, _ := client.ListHostedConfigurationVersions(context.TODO(),
			&appconfig.ListHostedConfigurationVersionsInput{ApplicationId: application.Id, ConfigurationProfileId: item.Id},
			func(options *appconfig.Options) {})
		currentVersion := aws.Int32(1)
		for _, items := range configurationVersions.Items {
			if items.VersionNumber > *currentVersion {
				currentVersion = &items.VersionNumber
			}
		}
		
		if v, ok := versionsMap[*item.Id]; ok {
			if *currentVersion > v {
				versionsMap[*item.Id] = *currentVersion
				if appName, okk := applicationPipelineMap[*item.Id]; okk {
					pipelineClient.StartPipelineExecution(context.TODO(), &codepipeline.StartPipelineExecutionInput{
						Name: &appName,
					})
					log.Println("重启Pipeline, 名称: ", appName)
				}
			}
		} else {
			versionsMap[*item.Id] = *currentVersion
		}
		
	}
	
}

func filterBaseApplication(applications *appconfig.ListApplicationsOutput, client *appconfig.Client) {
	// 找到对应的根应用程序
	applicationId := ""
	for _, item := range applications.Items {
		if *item.Name == BaseApplicationName {
			applicationId = *item.Id
			break
		}
	}
	if applicationId == "" {
		return
	}
	// 读取
	profiles, profileError := client.ListConfigurationProfiles(context.TODO(), &appconfig.ListConfigurationProfilesInput{ApplicationId: &applicationId})
	if profileError != nil || profiles == nil {
		return
	}
	currentProfileId := ""
	for _, item := range profiles.Items {
		if *item.Name == BaseApplicationProfileName {
			currentProfileId = *item.Id
		}
	}
	if currentProfileId == "" {
		return
	}
	configurationVersions, configureError := client.ListHostedConfigurationVersions(context.TODO(),
		&appconfig.ListHostedConfigurationVersionsInput{ApplicationId: &applicationId, ConfigurationProfileId: &currentProfileId})
	if configureError != nil || configurationVersions == nil {
		return
	}
	currentVersion := aws.Int32(1)
	for _, item := range configurationVersions.Items {
		if item.VersionNumber > *currentVersion {
			currentVersion = &item.VersionNumber
		}
	}
	configuration, _ := client.GetHostedConfigurationVersion(context.TODO(),
		&appconfig.GetHostedConfigurationVersionInput{ApplicationId: &applicationId, ConfigurationProfileId: &currentProfileId, VersionNumber: *currentVersion})
	json.Unmarshal(configuration.Content, &applicationPipelineMap)
}
