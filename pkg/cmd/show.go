// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewShowCmd returns a new show command.
func NewShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show (operator|gardener-dashboard|api|scheduler|controller-manager|etcd-operator|etcd-main|etcd-events|addon-manager|vpn-seed|vpn-shoot|machine-controller-manager|kubernetes-dashboard|prometheus|grafana|alertmanager|tf (infra|dns|ingress)|cluster-autoscaler)",
		Short: `Show details about endpoint/service and open in default browser if applicable`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 || len(args) > 2 {
				fmt.Println("Command must be in the format: show (operator|gardener-dashboard|api|scheduler|controller-manager|etcd-operator|etcd-main|etcd-events|addon-manager|vpn-seed|vpn-shoot|machine-controller-manager|kubernetes-dashboard|prometheus|grafana|alertmanager|tf (infra|dns|ingress)|cluster-autoscaler)")
				os.Exit(2)
			}
			var t Target
			targetFile, err := ioutil.ReadFile(pathTarget)
			checkError(err)
			err = yaml.Unmarshal(targetFile, &t)
			checkError(err)
			if len(t.Target) < 3 && (args[0] != "operator") && (args[0] != "tf") && (args[0] != "kubernetes-dashboard") && (args[0] != "etcd-operator") && (args[0] != "kibana") {
				fmt.Println("No shoot targeted")
				os.Exit(2)
			} else if (len(t.Target) < 2 && (args[0] == "tf")) || len(t.Target) < 3 && (args[0] == "tf") && (t.Target[1].Kind != "seed") || (len(t.Target) < 2 && (args[0] == "kibana")) {
				fmt.Println("No seed or shoot targeted")
				os.Exit(2)
			} else if len(t.Target) == 0 {
				fmt.Println("Target stack is empty")
				os.Exit(2)
			}
			switch args[0] {
			case "operator":
				showOperator()
			case "gardener-dashboard":
				showGardenerDashboard()
			case "api":
				showAPIServer()
			case "scheduler":
				showScheduler()
			case "controller-manager":
				showControllerManager()
			case "etcd-operator":
				showEtcdOperator()
			case "etcd-main":
				showEtcdMain()
			case "etcd-events":
				showEtcdEvents()
			case "addon-manager":
				showAddonManager()
			case "vpn-seed":
				showVpnSeed()
			case "vpn-shoot":
				showVpnShoot()
			case "machine-controller-manager":
				showMachineControllerManager()
			case "kubernetes-dashboard":
				showKubernetesDashboard()
			case "prometheus":
				showPrometheus()
			case "grafana":
				showGrafana()
			case "alertmanager":
				showAltermanager()
			case "kibana":
				showKibana()
			case "tf":
				if len(args) == 1 {
					showTf()
					break
				}
				switch args[1] {
				case "infra":
					showInfra()
				case "dns":
					showDNS()
				case "ingress":
					showIngress()
				default:
					fmt.Println("Command must be in the format: show (operator|gardener-dashboard|api|scheduler|controller-manager|etcd-operator|etcd-main|etcd-events|addon-manager|vpn-seed|vpn-shoot|machine-controller-manager|kubernetes-dashboard|prometheus|grafana|alertmanager|tf (infra|dns|ingress)|cluster-autoscaler)")
				}
			case "cluster-autoscaler":
				showClusterAutoscaler()
			default:
				fmt.Println("Command must be in the format: show (operator|gardener-dashboard|api|scheduler|controller-manager|etcd-operator|etcd-main|etcd-events|addon-manager|vpn-seed|vpn-shoot|machine-controller-manager|kubernetes-dashboard|prometheus|grafana|alertmanager|tf (infra|dns|ingress)|cluster-autoscaler)")
			}
		},
		ValidArgs: []string{"operator", "gardener-dashboard", "api", "scheduler", "controller-manager", "etcd-operator", "etcd-main", "etcd-events", "addon-manager", "vpn-seed", "vpn-shoot", "machine-controller-manager", "kubernetes-dashboard", "prometheus", "grafana", "alertmanager", "tf", "cluster-autoscaler"},
	}
}

// showPodGarden
func showPodGarden(podName string, namespace string) {
	var err error
	Client, err = clientToTarget("garden")
	checkError(err)
	pods, err := Client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	checkError(err)
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, podName) {
			err := ExecCmd(nil, "kubectl get pods "+pod.Name+" -o wide -n "+namespace, false, "KUBECONFIG="+KUBECONFIG)
			checkError(err)
		}
	}
}

// showOperator shows the garden operator pod in the garden cluster
func showOperator() {
	showPodGarden("gardener-apiserver", "garden")
	showPodGarden("gardener-controller-manager", "garden")
}

// showUI opens the gardener landing page
func showGardenerDashboard() {
	showPodGarden("gardener-dashboard", "garden")
	output, err := ExecCmdReturnOutput("bash", "-c", "export KUBECONFIG="+KUBECONFIG+"; kubectl get ingress gardener-dashboard-ingress -n garden")
	if err != nil {
		fmt.Println("Cmd was unsuccessful")
		os.Exit(2)
	}
	list := strings.Split(output, " ")
	url := "-"
	for _, val := range list {
		if strings.HasPrefix(val, "dashboard.") {
			url = val
		}
	}
	urls := strings.Split(url, ",")
	var filteredUrls []string
	match := false
	opened := false
	for index, url := range urls {
		for _, u := range filteredUrls {
			if url == u {
				match = true
			}
		}
		if !match {
			filteredUrls = append(filteredUrls, url)
			fmt.Println("URL-" + strconv.Itoa(index+1) + ": " + "https://" + url)
			if !opened {
				err := browser.OpenURL("https://" + url)
				checkError(err)
				opened = true
			}
		}
	}
}

// showPod is an abstraction to show pods in seed cluster controlplane or kube-system namespace of shoot
func showPod(toMatch string, toTarget TargetKind) {
	var target Target
	ReadTarget(pathTarget, &target)

	var namespace string
	if len(target.Target) == 2 {
		namespace = "garden"
	} else if len(target.Target) == 3 {
		namespace = getSeedNamespaceNameForShoot(target.Target[2].Name)
	}

	var err error
	Client, err = clientToTarget("seed")
	checkError(err)
	if toTarget == TargetKindShoot {
		namespace = "kube-system"
		Client, err = clientToTarget(TargetKindShoot)
		checkError(err)
	}
	pods, err := Client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	checkError(err)
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, toMatch) {
			err := ExecCmd(nil, "kubectl get pods "+pod.Name+" -o wide -n "+namespace, false, "KUBECONFIG="+KUBECONFIG)
			checkError(err)
		}
	}
}

// showAPIServer shows the pod for the api-server running in the targeted seed cluster
func showAPIServer() {
	showPod("kube-apiserver", "seed")
}

// showScheduler shows the pod for the running scheduler in the targeted seed cluster
func showScheduler() {
	showPod("kube-scheduler", "seed")
}

// showControllerManager shows the pod for the running controller-manager in the targeted seed cluster
func showControllerManager() {
	showPod("kube-controller-manager", "seed")
}

// showEtcdOperator shows the pod for the running etcd-operator in the targeted garden cluster
func showEtcdOperator() {
	showPodGarden("etcd-operator", "kube-system")
}

// showEtcdMain shows the pod for the running etcd-main in the targeted seed cluster
func showEtcdMain() {
	showPod("etcd-main", "seed")
}

// showEtcdEvents shows the pod for the running etcd-events in the targeted seed cluster
func showEtcdEvents() {
	showPod("etcd-events", "seed")
}

// showAddonManager shows the pod for the running addon-manager in the targeted seed cluster
func showAddonManager() {
	showPod("kube-addon-manager", "seed")
}

// showVpnSeed shows the pod for the running vpn-seed in the targeted seed cluster
func showVpnSeed() {
	showPod("kube-apiserver", "seed")
	showPod("prometheus-0", "seed")
}

// showVpnShoot shows the pod for the running vpn-shoot in the targeted shoot cluster
func showVpnShoot() {
	showPod("vpn-shoot", "shoot")
}

// showPrometheus shows the prometheus pod in the targeted seed cluster
func showPrometheus() {
	username, password = getMonitoringCredentials()
	showPod("prometheus", "seed")
	output, err := ExecCmdReturnOutput("bash", "-c", "export KUBECONFIG="+KUBECONFIG+"; kubectl get ingress prometheus -n "+getShootClusterName())
	if err != nil {
		fmt.Println("Cmd was unsuccessful")
		os.Exit(2)
	}
	list := strings.Split(output, " ")
	url := "-"
	for _, val := range list {
		if strings.HasPrefix(val, "p.") {
			url = val
		}
	}
	url = "https://" + username + ":" + password + "@" + url
	fmt.Println("URL: " + url)
	err = browser.OpenURL(url)
	checkError(err)
}

// showAltermanager shows the prometheus pods in the targeted seed cluster
func showAltermanager() {
	username, password = getMonitoringCredentials()
	showPod("alertmanager", "seed")
	output, err := ExecCmdReturnOutput("bash", "-c", "export KUBECONFIG="+KUBECONFIG+"; kubectl get ingress alertmanager -n "+getShootClusterName())
	if err != nil {
		fmt.Println("Cmd was unsuccessful")
		os.Exit(2)
	}
	list := strings.Split(output, " ")
	url := "-"
	for _, val := range list {
		if strings.HasPrefix(val, "a.") {
			url = val
		}
	}
	url = "https://" + username + ":" + password + "@" + url
	fmt.Println("URL: " + url)
	err = browser.OpenURL(url)
	checkError(err)
}

// showMachineControllerManager shows the prometheus pods in the targeted seed cluster
func showMachineControllerManager() {
	showPod("machine-controller-manager", "seed")
}

// showKubernetesDashboard shows the kubernetes dashboard for the targeted cluster
func showKubernetesDashboard() {
	var target Target
	ReadTarget(pathTarget, &target)
	gardenName := target.Stack()[0].Name
	if len(target.Target) == 1 {
		var err error
		Client, err = clientToTarget("garden")
		checkError(err)
		pods, err := Client.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
		checkError(err)
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "kubernetes-dashboard") {
				err := ExecCmd(nil, "kubectl get pods "+pod.Name+" -o wide -n kube-system", false, "KUBECONFIG="+KUBECONFIG)
				checkError(err)
			}
		}
	} else if len(target.Target) == 2 {
		namespace := "kube-system"
		if len(target.Target) == 2 && target.Target[1].Kind == "seed" {
			KUBECONFIG = filepath.Join(pathGardenHome, "cache", gardenName, "seeds", target.Target[1].Name, "kubeconfig.yaml")
		} else if len(target.Target) == 2 && target.Target[1].Kind == "project" {
			fmt.Println("Project targeted")
			os.Exit(2)
		}
		config, err := clientcmd.BuildConfigFromFlags("", KUBECONFIG)
		checkError(err)
		Client, err := kubernetes.NewForConfig(config)
		checkError(err)
		pods, err := Client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		checkError(err)
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "kubernetes-dashboard") {
				err := ExecCmd(nil, "kubectl get pods "+pod.Name+" -o wide -n "+namespace, false, "KUBECONFIG="+KUBECONFIG)
				checkError(err)
			}
		}
	} else if len(target.Target) == 3 {
		showPod("kubernetes-dashboard", "shoot")
	} else if len(target.Target) == 0 {
		fmt.Println("No target")
		os.Exit(2)
	}
	url := "http://127.0.0.1:8002/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/"
	err := browser.OpenURL(url)
	checkError(err)
	err = ExecCmd(nil, "kubectl proxy -p 8002", false, "KUBECONFIG="+KUBECONFIG)
	checkError(err)
}

// showGrafana shows the grafana dashboard for the targeted cluster
func showGrafana() {
	username, password = getMonitoringCredentials()
	showPod("grafana", "seed")
	output, err := ExecCmdReturnOutput("bash", "-c", "export KUBECONFIG="+KUBECONFIG+"; kubectl get ingress grafana -n "+getShootClusterName())
	if err != nil {
		fmt.Println("Cmd was unsuccessful")
		os.Exit(2)
	}
	list := strings.Split(output, " ")
	url := "-"
	for _, val := range list {
		if strings.HasPrefix(val, "g.") {
			url = val
		}
	}
	url = "https://" + username + ":" + password + "@" + url
	fmt.Println("URL: " + url)
	err = browser.OpenURL(url)
	checkError(err)
}

// showKibana shows the kibana dashboard for the targeted cluster
func showKibana() {
	username, password = getLoggingCredentials()
	showPod("kibana", "seed")

	var target Target
	ReadTarget(pathTarget, &target)

	var namespace string
	if len(target.Target) == 2 {
		namespace = "garden"
	} else if len(target.Target) == 3 {
		namespace = getShootClusterName()
	}

	output, err := ExecCmdReturnOutput("bash", "-c", "export KUBECONFIG="+KUBECONFIG+"; kubectl get ingress kibana -n "+namespace)
	if err != nil {
		fmt.Println("Cmd was unsuccessful")
		os.Exit(2)
	}
	list := strings.Split(output, " ")
	url := "-"
	for _, val := range list {
		if strings.HasPrefix(val, "k.") {
			url = val
		}
	}
	url = "https://" + username + ":" + password + "@" + url
	fmt.Println("URL: " + url)
	err = browser.OpenURL(url)
	checkError(err)
}

// showTerraform pods for specified name
func showTerraform(name string) {
	var err error
	Client, err = clientToTarget("seed")
	checkError(err)
	pods, err := Client.CoreV1().Pods("").List(metav1.ListOptions{})
	checkError(err)
	output := ""
	count := 0
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, name) && pod.Status.Phase == "Running" {
			output, err = ExecCmdReturnOutput("bash", "-c", "export KUBECONFIG="+KUBECONFIG+"; kubectl get pods "+pod.Name+" -o wide -n "+pod.Namespace)
			if err != nil {
				fmt.Println("Cmd was unsuccessful")
				os.Exit(2)
			}
			if count != 0 {
				fmt.Printf("%s\n", strings.Split(output, "\n")[1])
			} else {
				fmt.Printf("%s", output)
				count++
			}
		}
	}
}

// showTf shows the currently running infra tf-pods
func showTf() {
	showTerraform(".tf-job")
}

// showInfra shows the currently running infra tf-pods
func showInfra() {
	showTerraform(".infra.tf-job")
}

// showDNS shows the currently running dns tf-pods
func showDNS() {
	showTerraform(".dns.tf-job")
}

// showIngress shows the currently running ingress tf-pods
func showIngress() {
	showTerraform(".ingress.tf-job")
}

// showClusterAutoscaler shows the pod for the running cluster-autoscaler in the targeted seed cluster
func showClusterAutoscaler() {
	showPod("cluster-autoscaler", "seed")
}
