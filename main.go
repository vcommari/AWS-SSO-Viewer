package main

import (
    "context"
    "log"
	"fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"github.com/spf13/viper"
)

type Account struct {
    Name string
    Id string
}

type Group struct {
	Name string
	Id string
}

type PermissionSet struct {
	Name string
	Arn string
}

type AccountAssociation struct {
	AccountId string
	Group Group
	PermissionSet PermissionSet
}

type PermissionSetDetails struct {
	Name string
	Description string
	Arn string
}

func listAccounts(c *gin.Context) {
	var accountList []Account

    cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion(viper.GetString("region")),
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

				a := Account{*account.Name,*account.Id}
				accountList = append(accountList, a)
			}
		}
		nextToken = list.NextToken
	}
	c.JSON(http.StatusOK, accountList)
}

func listPSs(c *gin.Context) {
	PSList := new([]PermissionSetDetails)
	var wg sync.WaitGroup
	instanceArn := viper.GetString("instanceArn")

    cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion(viper.GetString("region")),
   	)
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }
	ssoadm := ssoadmin.NewFromConfig(cfg)
	
	nextToken := new(string)
	for nextToken != nil {
		list := new(ssoadmin.ListPermissionSetsOutput)
		if *nextToken == "" {
			list, err = ssoadm.ListPermissionSets(context.TODO(), &ssoadmin.ListPermissionSetsInput {
				InstanceArn : &instanceArn,
				MaxResults : aws.Int32(100),
			})
		} else {
			list, err = ssoadm.ListPermissionSets(context.TODO(), &ssoadmin.ListPermissionSetsInput {
				InstanceArn : &instanceArn,
				MaxResults : aws.Int32(100),
				NextToken : nextToken,
			})
		}
		if err != nil {
			log.Fatalf("failed to list pss, %v", err)
		}
		for _, ps := range list.PermissionSets {
			wg.Add(1)
			go func(PSList *[]PermissionSetDetails, arn string) {				
				defer wg.Done()
				psd := permissionSetDetailsFromArn(arn)
				*PSList = append (*PSList, psd)
			}(PSList, ps)
		}
		nextToken = list.NextToken
	}
	wg.Wait()
	c.JSON(http.StatusOK, PSList)
}

func permissionSetDetailsFromArn(PermissionSetArn string) PermissionSetDetails {
	PermissionSetDetails := new(PermissionSetDetails)
	instanceArn := viper.GetString("instanceArn")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion(viper.GetString("region")),
	)
	ssoadm := ssoadmin.NewFromConfig(cfg)
	perm, err := ssoadm.DescribePermissionSet(context.TODO(), &ssoadmin.DescribePermissionSetInput   {
		InstanceArn : &instanceArn,
		PermissionSetArn : &PermissionSetArn,
	})
	if err != nil {
		log.Fatalf("failed to describe permission set, %v", err)
	}
	PermissionSetDetails.Name = *perm.PermissionSet.Name
	if perm.PermissionSet.Description != nil {
		PermissionSetDetails.Description = *perm.PermissionSet.Description
	}
	PermissionSetDetails.Arn = PermissionSetArn
	return *PermissionSetDetails
}

func permissionSetNameFromArn(PermissionSetArn string) string {
	instanceArn := viper.GetString("instanceArn")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion(viper.GetString("region")),
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
	config.WithRegion(viper.GetString("region")),
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

func computePermissionSet(permissionset string, result *[]AccountAssociation, id string, host string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
	config.WithRegion(viper.GetString("region")),
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
			group := Group{principalName, *assigment.PrincipalId}
			permissionset := PermissionSet{permissionSetNameFromArn(*assigment.PermissionSetArn), *assigment.PermissionSetArn}
			a := AccountAssociation{id, group, permissionset}
			*result = append(*result, a)
		}
		nextToken = assignments.NextToken
	}
}

func computePermissionSetsList(permissionList []string, result *[]AccountAssociation, id string, host string) {
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
	result := new([]AccountAssociation)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion(viper.GetString("region")),
   	)
	if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }
	ssoadm := ssoadmin.NewFromConfig(cfg)
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
		computePermissionSetsList(permlist.PermissionSets, result, id, host)
		nextToken = permlist.NextToken
	}
	c.JSON(http.StatusOK, result)
}

func getPSPoliciesByARN(c *gin.Context) {
	arn := c.Request.URL.Query()["arn"][0]
	resultmap := make(map[string]string)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
   		config.WithRegion(viper.GetString("region")),
   	)
	if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }
	ssoadm := ssoadmin.NewFromConfig(cfg)
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
		for _,policy := range policieslist.AttachedManagedPolicies {
			resultmap[*policy.Name] = *policy.Arn
		}
		nextToken = policieslist.NextToken
	}
	c.JSON(http.StatusOK, resultmap)
}

func getPSInlineByARN(c *gin.Context) {
	arn := c.Request.URL.Query()["arn"][0]
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(viper.GetString("region")),
	)
	if err != nil {
 		log.Fatalf("unable to load SDK config, %v", err)
	}
	ssoadm := ssoadmin.NewFromConfig(cfg)
	instanceArn := viper.GetString("instanceArn")
	inlinePolicy, err := ssoadm.GetInlinePolicyForPermissionSet(context.TODO(), &ssoadmin.GetInlinePolicyForPermissionSetInput {
		InstanceArn : &instanceArn,
		PermissionSetArn : &arn,
	})
	c.String(http.StatusOK, *inlinePolicy.InlinePolicy)
}

func main() {
	// Set the file name of the configurations file
	viper.SetConfigName("aws-sso-viewer")
	// Set the path to look for the configurations file
	viper.AddConfigPath("/etc/")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	router := gin.Default()
	router.LoadHTMLFiles("staticfiles/index.html")
	router.Static("/staticfiles", "./staticfiles")
	router.StaticFile("/table.js", "./staticfiles/table.js")
	router.StaticFile("/styles.css", "./staticfiles/styles.css")
	router.StaticFile("/favicon.ico", "./staticfiles/favicon.ico")
	router.StaticFile("/SSO_logo.png", "./staticfiles/SSO_logo.png")
	router.StaticFile("/images/searchicon.png", "./staticfiles/searchicon.png")

	router.GET("/getaccount/:id", getPermissionsByAccountID)
	router.GET("/getpspolicies", getPSPoliciesByARN)
	router.GET("/getpsinline", getPSInlineByARN)
	router.GET("/accountslist", listAccounts)
	router.GET("/psslist", listPSs)
	
	router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "SSO Viewer",
		})
    })

	router.Run("0.0.0.0:" + viper.GetString("port"))
}


