package main

import (
    "context"
    "fmt"
    "log"
	//"encoding/json"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"strings"
	//c "./config"
	"github.com/spf13/viper"
	//"reflect"
)

type Account struct {
    Name string
    Id string
}

type AccountAssociation struct {
	AccountId string
	PermissionSetName string
	GroupName string
}

func listAccounts(c *gin.Context) {
	host := c.Request.Host
	var accountList []Account

    cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion("us-east-1"),
   	)
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    organization := organizations.NewFromConfig(cfg)
	
	nextToken := new(string)
	for nextToken != nil {
		list := new(organizations.ListAccountsOutput)
		if *nextToken == "" {
			list, err = organization.ListAccounts(context.TODO(), &organizations.ListAccountsInput{
				MaxResults : aws.Int32(10),
			})
		} else {
			list, err = organization.ListAccounts(context.TODO(), &organizations.ListAccountsInput{
				MaxResults : aws.Int32(10),
				NextToken : nextToken,
			})
		}
		if err != nil {
			log.Fatalf("failed to list accounts, %v", err)
		}
		for _, account := range list.Accounts {
			if account.Status == "ACTIVE" {
				url := "<a href=\"http://" + host + "/?page=accounts&?account=" + *account.Id + "\">" + *account.Name + "</a>"
				a := Account{url,*account.Id}
				accountList = append(accountList, a)
			}
		}
		nextToken = list.NextToken
	}
	c.JSON(http.StatusOK, accountList)
}

func permissionSetNameFromArn(PermissionSetArn string) string {
	instanceArn := viper.GetString("instanceArn")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion("us-east-1"),
	)
	ssoadm := ssoadmin.NewFromConfig(cfg)
	perm, err := ssoadm.DescribePermissionSet(context.TODO(), &ssoadmin.DescribePermissionSetInput   {
		InstanceArn : &instanceArn,
		PermissionSetArn : &PermissionSetArn,
	})
	if err != nil {
		log.Fatalf("failed to describe permission set, %v", err)
	}
	return *perm.PermissionSet.Name
}

func principalNameFromId(PrincipalId string, PrincipalType string) string {
	identityStoreId := viper.GetString("identityStoreId")
	cfg, _ := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion("us-east-1"),
	)
	myidentitystore := identitystore.NewFromConfig(cfg)
	if PrincipalType == "GROUP" {
		identity, err := myidentitystore.DescribeGroup(context.TODO(), &identitystore.DescribeGroupInput {
			GroupId : &PrincipalId,
			IdentityStoreId : &identityStoreId,
		})
		if err != nil {
			log.Fatalf("failed to describe group, %v", err)
		}
		return *identity.DisplayName 
	} else {
		identity, err := myidentitystore.DescribeUser(context.TODO(), &identitystore.DescribeUserInput {
			UserId : &PrincipalId,
			IdentityStoreId : &identityStoreId,
		})
		if err != nil {
			log.Fatalf("failed to describe user	#, %v", err)
		}
		return *identity.UserName 
	}
}

func computePermissionSet(permissionset string, result map[string]string, id string, host string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion("us-east-1"),
	)
	ssoadm := ssoadmin.NewFromConfig(cfg)
	instanceArn := viper.GetString("instanceArn")
	nextToken := new(string)
	for nextToken != nil {
		assignments:= new(ssoadmin.ListAccountAssignmentsOutput)
		if *nextToken == "" {
			assignments, err = ssoadm.ListAccountAssignments(context.TODO(), &ssoadmin.ListAccountAssignmentsInput  {
				InstanceArn : &instanceArn,
				AccountId : &id,
				PermissionSetArn : &permissionset,
				MaxResults : aws.Int32(10),
			})
		} else {
			assignments, err = ssoadm.ListAccountAssignments(context.TODO(), &ssoadmin.ListAccountAssignmentsInput  {
				InstanceArn : &instanceArn,
				AccountId : &id,
				PermissionSetArn : &permissionset,
				MaxResults : aws.Int32(10),
				NextToken : nextToken,
			})
		}
		if err != nil {
			log.Fatalf("failed to list accounts, %v", err)
		}
		for _, assigment :=  range assignments.AccountAssignments {
			var principalName string
			if assigment.PrincipalType == "GROUP" {
				principalName = principalNameFromId(*assigment.PrincipalId, "GROUP")
			} else {
				principalName = principalNameFromId(*assigment.PrincipalId, "USER")
			}
			permarn := permissionSetNameFromArn(*assigment.PermissionSetArn)
			permarn = strings.Replace(permarn, ":", "%3A", -1)
			permarn = strings.Replace(permarn, "/", "%2F", -1)
			permissionSetName := "<a href=\"http://" + host + "/?page=ps&arn=" +
			*assigment.PermissionSetArn + "\">" + permarn + "</a>"
			if _, ok := result[principalName]; ok {
				result[principalName] = result[principalName] + ", " + permissionSetName
			} else {
				result[principalName] = permissionSetName
			}
		}
		nextToken = assignments.NextToken
	}
}

func computePermissionSetsList(permissionList []string, result map[string]string, id string, host string) {
	//Takes a list of Permission set, converts the Id/Arns into Names and add it to the permissionList
	var wg sync.WaitGroup
	for _,perm := range permissionList {
		wg.Add(1)
		go func(permi string) {
            defer wg.Done()
			computePermissionSet(permi, result, id, host)
        }(perm)
	}
	wg.Wait()
}

func getPermissionsByAccountID(c *gin.Context) {
	id := c.Param("id")
	host := c.Request.Host
	//result := new([]AccountAssociation)
	resultmap := make(map[string]string)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion("us-east-1"),
   	)
	if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }
	ssoadm := ssoadmin.NewFromConfig(cfg)
	// TODO : Find a way to remove hardcoded arn
	instanceArn := viper.GetString("instanceArn")

	nextToken := new(string)
	for nextToken != nil {
		permlist := new(ssoadmin.ListPermissionSetsProvisionedToAccountOutput)
		if *nextToken == "" {
			permlist, err = ssoadm.ListPermissionSetsProvisionedToAccount(context.TODO(), &ssoadmin.ListPermissionSetsProvisionedToAccountInput{
				InstanceArn : &instanceArn,
				AccountId : &id,
				MaxResults : aws.Int32(10),
			})
		} else {
			permlist, err = ssoadm.ListPermissionSetsProvisionedToAccount(context.TODO(), &ssoadmin.ListPermissionSetsProvisionedToAccountInput {
				InstanceArn : &instanceArn,
				AccountId : &id,
				MaxResults : aws.Int32(10),
				NextToken : nextToken,
			})
		}
		if err != nil {
			log.Fatalf("failed to list accounts, %v", err)
		}
		computePermissionSetsList(permlist.PermissionSets, resultmap, id, host)
		nextToken = permlist.NextToken
	}
	fmt.Println(resultmap)
	c.JSON(http.StatusOK, resultmap)
}

func getPSPoliciesByARN(c *gin.Context) {
	arn := c.Request.URL.Query()["arn"][0]
	resultmap := make(map[string]string)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion("us-east-1"),
   	)
	if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }
	ssoadm := ssoadmin.NewFromConfig(cfg)
	// TODO : Find a way to remove hardcoded arn
	instanceArn := viper.GetString("instanceArn")
	nextToken := new(string)
	for nextToken != nil {
		policieslist := new(ssoadmin.ListManagedPoliciesInPermissionSetOutput)
		if *nextToken == "" {
			policieslist, err = ssoadm.ListManagedPoliciesInPermissionSet(context.TODO(), &ssoadmin.ListManagedPoliciesInPermissionSetInput {
				InstanceArn : &instanceArn,
				PermissionSetArn : &arn,
				MaxResults : aws.Int32(10),
			})
		} else {
			policieslist, err = ssoadm.ListManagedPoliciesInPermissionSet(context.TODO(), &ssoadmin.ListManagedPoliciesInPermissionSetInput {
				InstanceArn : &instanceArn,
				PermissionSetArn : &arn,
				MaxResults : aws.Int32(10),
				NextToken : nextToken,
			})
		}
		if err != nil {
			log.Fatalf("failed to list policies, %v", err)
		}
		//computePermissionSetsList(permlist.PermissionSets, resultmap, id, host)
		for _,policy := range policieslist.AttachedManagedPolicies {
			resultmap[*policy.Name] = *policy.Arn
			fmt.Println(*policy.Name)
		}
		//fmt.Println(policieslist.AttachedManagedPolicies["Name"])
		nextToken = policieslist.NextToken
	}
	c.JSON(http.StatusOK, resultmap)
}

func getPSInlineByARN(c *gin.Context) {
	arn := c.Request.URL.Query()["arn"][0]
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
 		log.Fatalf("unable to load SDK config, %v", err)
	}
	ssoadm := ssoadmin.NewFromConfig(cfg)
	// TODO : Find a way to remove hardcoded arn
	instanceArn := viper.GetString("instanceArn")
	inlinePolicy, err := ssoadm.GetInlinePolicyForPermissionSet(context.TODO(), &ssoadmin.GetInlinePolicyForPermissionSetInput {
		InstanceArn : &instanceArn,
		PermissionSetArn : &arn,
	})
	c.String(http.StatusOK, *inlinePolicy.InlinePolicy)
}

func main() {
	// Set the file name of the configurations file
	viper.SetConfigName("config")
	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	router := gin.Default()
	router.LoadHTMLFiles("staticfiles/index.html")
	//router.LoadHTMLFiles("staticfiles/account.html")
	router.Static("/staticfiles", "./staticfiles")
	router.StaticFile("/table.js", "./staticfiles/table.js")
	router.StaticFile("/styles.css", "./staticfiles/styles.css")
	router.StaticFile("/favicon.ico", "./staticfiles/favicon.ico")
	router.StaticFile("/images/searchicon.png", "./staticfiles/searchicon.png")

	//router.GET("/getusers/:group", getUsersByGroupID) 
	router.GET("/getaccount/:id", getPermissionsByAccountID)
	router.GET("/getpspolicies", getPSPoliciesByARN)
	router.GET("/getpsinline", getPSInlineByARN)
	router.GET("/accountslist", listAccounts)
	
	router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "SSO Viewer",
		})
    })

	router.Run("localhost:" + viper.GetString("port"))
}


