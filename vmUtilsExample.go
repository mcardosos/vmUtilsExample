package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/hostedservice"
	"github.com/Azure/azure-sdk-for-go/management/storageservice"
	"github.com/Azure/azure-sdk-for-go/management/virtualmachine"
	"github.com/Azure/azure-sdk-for-go/management/vmutils"
)

func main() {
	dnsName := "test-vm-from-go"
	storageAccount := fmt.Sprintf("stoacc%v", randomString(5))
	location := "central us"
	vmSize := "Small"
	vmImage := "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04-LTS-amd64-server-20140724-en-us-30GB"
	userName := "testuser"
	userPassword := "Test123"

	fmt.Println("Create client...")
	client, err := management.ClientFromPublishSettingsFile("credentials.publishsettings", "")
	if err != nil {
		panic(err)
	}

	fmt.Println("Create hosted service...")
	// create hosted service
	if err := hostedservice.NewClient(client).CreateHostedService(hostedservice.CreateHostedServiceParameters{
		ServiceName: dnsName,
		Location:    location,
		Label:       base64.StdEncoding.EncodeToString([]byte(dnsName))}); err != nil {
		panic(err)
	}

	fmt.Println("Create storage account...")
	_, err = storageservice.NewClient(client).CreateStorageService(storageservice.StorageAccountCreateParameters{
		ServiceName: storageAccount,
		Label:       base64.URLEncoding.EncodeToString([]byte(storageAccount)),
		Location:    location,
		AccountType: storageservice.AccountTypeStandardLRS})
	if err != nil {
		panic(err)
	}

	fmt.Println("Create VM...")
	// create virtual machine
	role := vmutils.NewVMConfiguration(dnsName, vmSize)
	vmutils.ConfigureDeploymentFromPlatformImage(
		&role,
		vmImage,
		fmt.Sprintf("http://%s.blob.core.windows.net/sdktest/%s.vhd", storageAccount, dnsName),
		"")
	vmutils.ConfigureForLinux(&role, dnsName, userName, userPassword)
	vmutils.ConfigureWithPublicSSH(&role)

	fmt.Println("Deploy...")
	operationID, err := virtualmachine.NewClient(client).
		CreateDeployment(role, dnsName, virtualmachine.CreateDeploymentOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Waiting for deployment to finish...")
	if err := client.WaitForOperation(operationID, nil); err != nil {
		panic(err)
	}
}

func randomString(strLen int) string {
	ran := 'z' - 'a' + 10
	text := make([]byte, strLen)
	for i := range text {
		char := rand.Int()
		char %= int(ran)
		if char <= 9 {
			char += '0'
		} else {
			char += 'a' - 10
		}
		text[i] = byte(char)
	}
	return string(text)
}
