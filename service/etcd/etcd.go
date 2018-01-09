package etcd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/docktermj/go-logger/logger"
	errorWrap "github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Etcd struct {
	Context                context.Context
	ClientEndpoints        string
	ClusterClientEndpoints string
	EtcdClient             *clientv3.Client
	PeerEndpoints          string
	WaitGroup              *sync.WaitGroup
}

// Return an array of string without any duplicates.
// https://play.golang.org/p/GG2RQ1ot6t
func removeDuplicates(stringArray *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *stringArray {
		if !found[x] {
			found[x] = true
			(*stringArray)[j] = (*stringArray)[i]
			j++
		}
	}
	*stringArray = (*stringArray)[:j]
}

// Given an array of string of URLs, create an array of url.URL.
func createUrlListFromArrayOfString(urlStrings []string) []url.URL {
	var result []url.URL
	for _, urlString := range urlStrings {
		aUrl, err := url.Parse(urlString)
		if err == nil {
			result = append(result, *aUrl)
		} else {
			logger.Warnf("Could not parse URL: %s Reason: %v", urlString, err)
		}
	}
	return result
}

// Given a comma-deliminated string of URLs, create an array of url.URL.
func createUrlList(urlListString string) []url.URL {
	var result []url.URL
	if len(urlListString) > 0 {
		result = createUrlListFromArrayOfString(strings.Split(urlListString, ","))
	}
	return result
}

// Create a Etcd client.
func getEtcdClient(ctx context.Context, clusterClientEndpoints string) (*clientv3.Client, error) {

	// Normalize Cluster Client Endpoints.

	etcdClusterClientEndpoints := createUrlList(clusterClientEndpoints)
	if len(etcdClusterClientEndpoints) == 0 {
		return nil, nil
	}

	// Construct normalized endpoints for the Etcd Client.

	endpoints := []string{}
	for _, url := range etcdClusterClientEndpoints {
		endpoints = append(endpoints, fmt.Sprintf("%s:%s", url.Hostname(), url.Port()))
	}

	// Return a new EtcdClient.

	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		Context:     ctx,
	})
}

// Remove the current member from the cluster.
func removeMemberFromCluster(etcdClient *clientv3.Client) {
    ctx := context.Background()
	response, err := etcdClient.MemberList(ctx)
	if err != nil {
		panic(err)
	}
	id := response.Members[0].ID
	logger.Debugf("Deleting member %d\n", id)
	_, err = etcdClient.MemberRemove(ctx, id)
	if err != nil {
		panic(err)
	}
}

//
func addMemberToExistingCluster(ctx context.Context, peerEndpoints string, etcdClient *clientv3.Client) (func(), error) {

	// If Etcd client has not been created, simply return.

	if etcdClient == nil {
		return func() {
		}, nil
	}

	// Determine the members to add.

	etcdPeerEndpoints := createUrlList(peerEndpoints)
	peerAddrs := []string{}
	for _, url := range etcdPeerEndpoints {
		peerAddrs = append(peerAddrs, fmt.Sprintf("%s://%s:%s", url.Scheme, url.Hostname(), url.Port()))
	}

	// Perform the equivalent of "etcdctl member add".

	_, err := etcdClient.MemberAdd(ctx, peerAddrs)
	if err != nil {
		errString := fmt.Sprintf("etcdClient.MemberAdd(ctx, %s) failed. Err: %+v", peerAddrs, err)
		logger.Warn(errString)
		return func() {
		}, errorWrap.Wrap(err, errString)
	}

	return func() {
		removeMemberFromCluster(etcdClient)
	}, nil
}

func getInitialCluster(ctx context.Context, peerEndpoints string, etcdClient *clientv3.Client) (string, string, error) {
	result := ""
	resultClusterState := "new"
	fmtTemplate := "%s%s:%s=http://%s:%s,"

	// Add the new etcd instance(s) to initial cluster.

	urlArray := strings.Split(peerEndpoints, ",")

	// Add existing etcd instances to initial cluster.

	if etcdClient != nil {

		// Using the Etcd client, find the aggregate list of "peers".

		members, err := etcdClient.MemberList(ctx)
		if err != nil {
			errString := "etcdClient.MemberList() failed."
			logger.Warnf("%s %+v", errString, err)
			return result, resultClusterState, errorWrap.Wrap(err, errString)
		}
		for _, member := range members.Members {
			for _, peer := range member.GetPeerURLs() {
				urlArray = append(urlArray, peer)
			}
		}
		resultClusterState = "existing"
	}

	// De-duplicate and finalize the result.

	removeDuplicates(&urlArray)
	urls := createUrlListFromArrayOfString(urlArray)
	if len(urls) > 0 {
		for _, url := range urls {
			result = fmt.Sprintf(fmtTemplate, result, url.Hostname(), url.Port(), url.Hostname(), url.Port())
		}
	}

	return strings.TrimSuffix(result, ","), resultClusterState, nil
}

// Get configuration values.
// https://github.com/coreos/etcd/blob/master/embed/config.go
func getEtcdConfig(ctx context.Context, peerEndpoints string, clientEndpoints string, etcdClient *clientv3.Client) (*embed.Config, error) {
	result := embed.NewConfig()
	var err error

	// Configuration options.

	etcdClientEndpoints := createUrlList(clientEndpoints)
	etcdPeerEndpoints := createUrlList(peerEndpoints)

	// Construct embed.Config.

	result.Dir = fmt.Sprintf("%s:%s.etcd", etcdPeerEndpoints[0].Hostname(), etcdPeerEndpoints[0].Port())
	result.Name = fmt.Sprintf("%s:%s", etcdPeerEndpoints[0].Hostname(), etcdPeerEndpoints[0].Port())

	// Load URL lists.

	result.ACUrls = etcdClientEndpoints
	result.LCUrls = etcdClientEndpoints
	result.APUrls = etcdPeerEndpoints
	result.LPUrls = etcdPeerEndpoints
	result.InitialCluster, result.ClusterState, err = getInitialCluster(ctx, peerEndpoints, etcdClient)
	if err != nil {
		errString := fmt.Sprintf("getInitialCluster(%s) failed.", peerEndpoints)
		logger.Warnf("%s %+v", errString, err)
		return result, errorWrap.Wrap(err, errString)
	}

	logger.Debugf("etcd configuration: %+v", result)
	return result, nil
}

// ----------------------------------------------------------------------------
// Service with parameters interface
// ----------------------------------------------------------------------------

func (etcd Etcd) Run() error {

	// Synchronize the services at shutdown.

	if etcd.WaitGroup != nil {
		etcd.WaitGroup.Add(1)
		defer etcd.WaitGroup.Done()
	}

	// If not supplied, try to create an Etcd client.

	if etcd.EtcdClient == nil {
		etcdClient, err := getEtcdClient(etcd.Context, etcd.ClusterClientEndpoints)
		if err != nil {
			errString := fmt.Sprintf("getEtcdClient(ctx, %s) failed.", etcd.ClusterClientEndpoints)
			logger.Warnf("%s %+v", errString, err)
			return errorWrap.Wrap(err, errString)
		}
		if etcdClient != nil {
			defer etcdClient.Close()
			etcd.EtcdClient = etcdClient
		}
	}

	// Step 1: Add new member to existing cluster.
	// https://coreos.com/etcd/docs/latest/op-guide/runtime-configuration.html#add-a-new-member

	removeFromExistingCluster, err := addMemberToExistingCluster(etcd.Context, etcd.PeerEndpoints, etcd.EtcdClient)
	if err != nil {
		errString := fmt.Sprintf("addMemberToExistingCluster(ctx, %s) failed.", etcd.PeerEndpoints)
		logger.Warnf("%s %+v", errString, err)
		return errorWrap.Wrap(err, errString)
	}
	defer removeFromExistingCluster()

	// Step 2: Start new member.
	// https://coreos.com/etcd/docs/latest/op-guide/runtime-configuration.html#add-a-new-member

	// Configuration

	etcdConfig, err := getEtcdConfig(etcd.Context, etcd.PeerEndpoints, etcd.ClientEndpoints, etcd.EtcdClient)
	if err != nil {
		errString := fmt.Sprintf("getEtcdConfig(ctx, %s, %s) failed.", etcd.PeerEndpoints, etcd.ClientEndpoints)
		logger.Warnf("%s %+v", errString, err)
		return errorWrap.Wrap(err, errString)
	}

	// Start Etcd server.

	etcdService, err := embed.StartEtcd(etcdConfig)
	if err != nil {
		errString := fmt.Sprintf("etcd server did not start with configuration: %+v.", etcdConfig)
		logger.Warnf("%s %+v", errString, err)
		return errorWrap.Wrap(err, errString)
	}
	defer etcdService.Close()

	// Monitor etcd service.

	select {
	case <-etcdService.Server.ReadyNotify():
		logger.Infof("etcd server is ready!")
	case <-time.After(60 * time.Second):
		etcdService.Server.Stop() // trigger a shutdown
		errString := "etcd server took too long to start."
		logger.Warnf(errString)
		return errors.New(errString)
	}

	// Run service until the context is done.

	<-etcd.Context.Done()

	// Epilog.

	logger.Debugf("Done\n")

	return nil
}

// ----------------------------------------------------------------------------
// Service interface
// ----------------------------------------------------------------------------

func Service(ctx context.Context) error {
	service := Etcd{
		Context:                ctx,
		ClientEndpoints:        viper.GetString("etcdClientEndpoints"),
		ClusterClientEndpoints: viper.GetString("etcdClusterClientEndpoints"),
		PeerEndpoints:          viper.GetString("etcdPeerEndpoints"),
	}
	return service.Run()
}

func ServiceWithWaitGroup(ctx context.Context, waitGroup *sync.WaitGroup) error {
	service := Etcd{
		Context:                ctx,
		ClientEndpoints:        viper.GetString("etcdClientEndpoints"),
		ClusterClientEndpoints: viper.GetString("etcdClusterClientEndpoints"),
		PeerEndpoints:          viper.GetString("etcdPeerEndpoints"),
		WaitGroup:              waitGroup,
	}
	return service.Run()
}
