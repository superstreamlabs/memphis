def repoUrlPrefix = "memphisos"
def imageName = "memphis-control-plane"
def gitURL = "git@github.com:Memphis-OS/memphis-control-plane.git"
def gitBranch = "master"
def versionTag = "master"
String unique_id = org.apache.commons.lang.RandomStringUtils.random(4, false, true)
def namespace = "memphis"
def test_suffix = "test"
//def DOCKER_HUB_CREDS = credentials('docker-hub')



node {
  try{
    stage('SCM checkout') {
        git credentialsId: 'main-github', url: gitURL, branch: gitBranch
    }

    stage('Login to Docker Hub') {
	    withCredentials([usernamePassword(credentialsId: 'docker-hub', usernameVariable: 'DOCKER_HUB_CREDS_USR', passwordVariable: 'DOCKER_HUB_CREDS_PSW')]) {
		  sh "docker login -u $DOCKER_HUB_CREDS_USR -p $DOCKER_HUB_CREDS_PSW"
	    }
    }

    stage('Create memphis namespace in Kubernetes'){
      sh "kubectl create namespace memphis-$unique_id --dry-run=client -o yaml | kubectl apply -f -"
      //sh "sleep 40"
    }

    stage('Build and push docker image to Docker Hub') {
	    sh "docker buildx build --push -t ${repoUrlPrefix}/${imageName}-${versionTag}-${test_suffix} ."
    }

    stage('Tests - Install/upgrade Memphis cli') {
      sh "sudo npm uninstall memphis-dev-cli"
      sh "sudo npm i memphis-dev-cli -g"
    }

    
    ////////////////////////////////////////
    //////////// Docker-Compose ////////////
    ////////////////////////////////////////

    stage('Tests - Docker compose install') {
      sh "rm -rf memphis-infra"
      sh "git clone git@github.com:Memphis-OS/memphis-infra.git"
      sh "docker-compose -f ./memphis-infra/${versionTag}/docker/docker-compose-dev-memphis-control-plane.yml -p memphis up -d"
    }
/*
    stage('Tests - Run e2e tests over Docker') {
      sh "rm -rf memphis-e2e-tests"
      sh "git clone git@github.com:Memphis-OS/memphis-e2e-tests.git"
      sh "npm install --prefix ./memphis-e2e-tests"
      sh "node ./memphis-e2e-tests/index.js docker"
    }*/

    stage('Tests - Remove Docker compose') {
      sh "docker-compose -f ./memphis-infra/${versionTag}/docker/docker-compose-dev-memphis-control-plane.yml -p memphis down"
    }

    ////////////////////////////////////////
    ////////////   Kubernetes   ////////////
    ////////////////////////////////////////

    stage('Tests - Install memphis with helm') {
      sh "helm install memphis-tests memphis-infra/${versionTag}/kubernetes/memphis --set analytics='false',teston='cp' --create-namespace --namespace memphis-$unique_id"
      sh 'sleep 40'
    }

    stage('Open port forwarding to memphis service') {
      sh "nohup kubectl port-forward service/memphis-ui 9000:80 --namespace memphis-$unique_id &"
      sh "sleep 5"
      sh "nohup kubectl port-forward service/memphis-cluster 7766:7766 6666:6666 5555:5555 --namespace memphis-$unique_id &"
      sh "sleep 5"
    }

/*
    stage('Tests - Run e2e tests over kubernetes') {
      //sh "npm install --prefix ./memphis-e2e-tests"
      sh "node ./memphis-e2e-tests/index.js kubernetes memphis-$unique_id"
    }*/

    stage('Tests - Uninstall helm') {
      sh "helm uninstall memphis-tests -n memphis-$unique_id"
      sh "kubectl delete ns memphis-$unique_id &"
      sh "lsof -i :5555,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh"
    }

    stage('Tests - Remove used directories') {
      sh "rm -rf memphis-infra"
      //sh "rm -rf memphis-e2e-tests"
    }


    ////////////////////////////////////////
    ////////////  Build & Push  ////////////
    ////////////////////////////////////////
/*
    stage('Build and push image to Docker Hub') {
      sh "docker buildx build --push -t ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 ."
    }*/

    ////////////////////////////////////////
    ////////////Test Public Repo////////////
    ////////////////////////////////////////

    ////////////////////////////////////////
    //////////// Docker-Compose ////////////
    ////////////////////////////////////////
/*
    stage('Tests - Docker compose install') {
      sh "rm -rf memphis-docker"
      sh "git clone git@github.com:Memphis-OS/memphis-docker.git"
      sh "docker-compose -f ./memphis-docker/docker-compose-dev.yml -p memphis up -d"
    }

    stage('Tests - Run e2e tests over Docker') {
      //sh "npm install --prefix ./memphis-e2e-tests"
      sh "node ./memphis-e2e-tests/index.js docker"
    }

    stage('Tests - Remove Docker compose') {
      sh "docker-compose -f ./memphis-docker/docker-compose-dev.yml -p memphis down"
      sh "rm -rf memphis-docker"
    }

    ////////////////////////////////////////
    ////////////   Kubernetes   ////////////
    ////////////////////////////////////////

    stage('Tests - Install memphis with helm') {
      sh "rm -rf memphis-k8s"
      sh "git clone git@github.com:Memphis-OS/memphis-k8s.git"
      sh "helm install memphis-tests memphis-k8s/memphis --set analytics='false' --create-namespace --namespace memphis-$unique_id"
      sh 'sleep 40'
    }

    stage('Open port forwarding to memphis service') {
      sh "nohup kubectl port-forward service/memphis-ui 9000:80 --namespace memphis-$unique_id &"
      sh "sleep 5"
      sh "nohup kubectl port-forward service/memphis-cluster 7766:7766 6666:6666 5555:5555 --namespace memphis-$unique_id &"
      sh "sleep 5"
    }


    stage('Tests - Run e2e tests over kubernetes') {
      //sh "npm install --prefix ./memphis-e2e-tests"
      sh "node ./memphis-e2e-tests/index.js kubernetes memphis-$unique_id"
    }

    stage('Tests - Uninstall helm') {
      sh "helm uninstall memphis-tests -n memphis-$unique_id"
      sh "kubectl delete ns memphis-$unique_id &"
      sh "lsof -i :5555,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh"
    }

    stage('Tests - Remove used directories') {
      sh "rm -rf memphis-k8s"
      sh "rm -rf memphis-e2e-tests"
    }*/
	  
	  
	  
    notifySuccessful()

 } catch (e) {
      currentBuild.result = "FAILED"
      sh "helm uninstall memphis-tests -n memphis-$unique_id"
      sh "kubectl delete ns memphis-$unique_id &"
      cleanWs()
      notifyFailed()
      throw e
  }
}

def notifySuccessful() {
  emailext (
      subject: "SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
      body: """<p>SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':</p>
        <p>Check console output at &QUOT;<a href='${env.BUILD_URL}'>${env.JOB_NAME} [${env.BUILD_NUMBER}]</a>&QUOT;</p>""",
      recipientProviders: [[$class: 'DevelopersRecipientProvider']]
    )
}

def notifyFailed() {
  emailext (
      subject: "FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
      body: """<p>FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':</p>
        <p>Check console output at &QUOT;<a href='${env.BUILD_URL}'>${env.JOB_NAME} [${env.BUILD_NUMBER}]</a>&QUOT;</p>""",
      recipientProviders: [[$class: 'DevelopersRecipientProvider']]
    )
}
