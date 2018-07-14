package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"agones.dev/agones/pkg/apis/stable/v1alpha1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	//informers "agones.dev/agones/pkg/client/informers/externalversions"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// use PORT environment variable, or default to 8080
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	// register hello function to handle all requests
	server := http.NewServeMux()
	server.HandleFunc("/", index)
	server.HandleFunc("/newgame", newGame)

	// start the web server on port and accept requests
	log.Printf("Server listening on port %s", port)
	err := http.ListenAndServe(":"+port, server)
	log.Fatal(err)
}

// hello responds to the request with a plain-text "Hello, world" message.
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, indexPage)
}

func newGame(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Could not create in cluster config: %v", err)
	}

	// Access to standard Kubernetes resources through the Kubernetes Clientset
	// We don't actually need this for this example, but it's just here for
	// illustrative purposes
	//kubeClient, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//		log.Fatalf("Could not create the kubernetes clientset: %v", err)
	//	}

	// Access to the Agones resources through the Agones Clientset
	// Note that we reuse the same config as we used for the Kubernetes Clientset
	agonesClient, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("Could not create the agones api clientset: %v", err)
	}

	// Create a GameServer
	gs := &v1alpha1.GameServer{ObjectMeta: metav1.ObjectMeta{GenerateName: "simple-ws-", Namespace: "default"},
		Spec: v1alpha1.GameServerSpec{
			PortPolicy:    "dynamic",
			Protocol:      "TCP",
			ContainerPort: 7654,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "simple-ws", Image: "gcr.io/YOUR_PROJECT_ID/simple-ws:latest"}},
				},
			},
		},
	}
	newGS, err := agonesClient.StableV1alpha1().GameServers("default").Create(gs)
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("Failed to create gameserver: %v", err))
		return
	}

	name := newGS.ObjectMeta.Name
	options := metav1.GetOptions{}
	for {
		gs, err := agonesClient.StableV1alpha1().GameServers("default").Get(name, options)
		if err != nil {
			fmt.Fprintf(w, fmt.Sprintf("Error updating gameserver: %v", err))
			return
		}
		switch gs.Status.State {
		case v1alpha1.Ready:
			t, err := template.New("serverPage").Parse(serverTemplate)
			if err != nil {
				fmt.Fprintf(w, fmt.Sprintf("Error creating template: %v", err))
				return
			}

			data := struct {
				Name string
				Host string
				Port int32
			}{
				Name: name,
				Host: gs.Status.Address,
				Port: gs.Status.Port,
			}

			err = t.Execute(w, data)
			if err != nil {
				fmt.Fprintf(w, fmt.Sprintf("Error rendering template: %v", err))
				return
			}
			return
		case v1alpha1.Error, v1alpha1.Unhealthy, v1alpha1.Shutdown:
			fmt.Fprintf(w, "Error creating gameserver.")
			return
		default:
			time.Sleep(time.Second * 5)
		}
	}
}

const indexPage = `
<!DOCTYPE html>
<html>
  <head>
  <title>Create New Server</title>
  </head>
  <body>
  <a href="/newgame">Create new server</a>
  </body>
</html>`

const serverTemplate = `
<!DOCTYPE html>
<html>
  <head>
  <title>{{.Name}}</title>
  </head>
  <body>
  <ul>
    <li>{{.Name}}</title>
	<li><a href="http://{{.Host}}:{{.Port}}">{{.Host}}:{{.Port}}</a></li>
  </ul>
  </body>
</html>`
