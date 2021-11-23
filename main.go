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
	var accountList []Account
	host := c.Request.Host

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
			url := "<a href=\"http://" + host + "/account/?account=" + *account.Id + "\">" + *account.Name + "</a>"
			a := Account{url,*account.Id}
			accountList = append(accountList, a)
		}
		nextToken = list.NextToken
	}
	c.JSON(http.StatusOK, accountList)
}

func permissionSetNameFromArn(PermissionSetArn string) string {
	instanceArn := "arn:aws:sso:::instance/ssoins-72231249bc307843"
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

func principalNameFromId(PrincipalId string) string {
	identityStoreId := "d-9067081d78"
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion("us-east-1"),
	)
	myidentitystore := identitystore.NewFromConfig(cfg)
	identity, err := myidentitystore.DescribeGroup(context.TODO(), &identitystore.DescribeGroupInput {
		GroupId : &PrincipalId,
		IdentityStoreId : &identityStoreId,
	})
	if err != nil {
		log.Fatalf("failed to describe group, %v", err)
	}
	return *identity.DisplayName 
}

func computePermissionSet(permissionset string, result *[]AccountAssociation, id string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion("us-east-1"),
	)
	ssoadm := ssoadmin.NewFromConfig(cfg)
	// TODO : Find a way to remove hardcoded arn
	instanceArn := "arn:aws:sso:::instance/ssoins-72231249bc307843"
	//fmt.Println(permissionset)
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
			principalName := principalNameFromId(*assigment.PrincipalId)
			permissionSetName := permissionSetNameFromArn(*assigment.PermissionSetArn)
			*result = append(*result, AccountAssociation{*assigment.AccountId, permissionSetName, principalName})
		}
		nextToken = assignments.NextToken
	}
}

func computePermissionSetsList(permissionList []string, result *[]AccountAssociation, id string) {
	//Takes a list of Permission set, converts the Id/Arns into Names and add it to the permissionList
	var wg sync.WaitGroup
	for _,perm := range permissionList {
		wg.Add(1)
		go func(permi string) {
            defer wg.Done()
			computePermissionSet(permi, result, id)
        }(perm)
	}
	wg.Wait()
}

func getPermissionsByAccountID(c *gin.Context) {
	id := c.Param("id")
	result := new([]AccountAssociation)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion("us-east-1"),
   	)
	if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }
	ssoadm := ssoadmin.NewFromConfig(cfg)
	// TODO : Find a way to remove hardcoded arn
	instanceArn := "arn:aws:sso:::instance/ssoins-72231249bc307843"

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
		//fmt.Println(permlist.PermissionSets)
		if err != nil {
			log.Fatalf("failed to list accounts, %v", err)
		}
		computePermissionSetsList(permlist.PermissionSets, result, id)
		nextToken = permlist.NextToken
	}
	//sortedresult := new ([]AccountAssociation)
	for _,association := range *result {
		fmt.Println(association.GroupName)
	}
	c.JSON(http.StatusOK, result)
}

func main() {
	router := gin.Default()
	router.LoadHTMLFiles("staticfiles/index.html")
	//router.LoadHTMLFiles("staticfiles/account.html")
	router.Static("/staticfiles", "./staticfiles")
	router.StaticFile("/table.js", "./staticfiles/table.js")
	router.StaticFile("/styles.css", "./staticfiles/styles.css")
	router.StaticFile("/favicon.ico", "./staticfiles/favicon.ico")
	router.StaticFile("/images/searchicon.png", "./staticfiles/searchicon.png")
	router.StaticFile("/account","staticfiles/account.html")

	router.GET("/getaccount/:id", getPermissionsByAccountID)
	router.GET("/accountslist", listAccounts)
	
	router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "SSO Viewer",
		})
    })

	router.Run("localhost:8080")
}


