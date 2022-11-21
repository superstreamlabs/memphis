def gitBranch = env.BRANCH_NAME
def imageName = "memphis-broker"
def gitURL = "git@github.com:Memphisdev/memphis-broker.git"
def repoUrlPrefix = "memphisos"
def test_suffix = "test"
String unique_id = org.apache.commons.lang.RandomStringUtils.random(4, false, true)
node {
  git credentialsId: 'main-github', url: gitURL, branch: gitBranch
  def versionTag = readFile "./version.conf"
  
  try{
	  
    stage('Login to Docker Hub') {
      withCredentials([usernamePassword(credentialsId: 'docker-hub', usernameVariable: 'DOCKER_HUB_CREDS_USR', passwordVariable: 'DOCKER_HUB_CREDS_PSW')]) {
      sh 'docker login -u $DOCKER_HUB_CREDS_USR -p $DOCKER_HUB_CREDS_PSW'
      }
    }

    stage('Create memphis namespace in Kubernetes'){
      sh "kubectl config use-context minikube"
      sh "kubectl create namespace memphis-$unique_id --dry-run=client -o yaml | kubectl apply -f -"
      sh "aws s3 cp s3://memphis-jenkins-backup-bucket/regcred.yaml ."
      sh "kubectl apply -f regcred.yaml -n memphis-$unique_id"
      sh "kubectl patch serviceaccount default -p '{\"imagePullSecrets\": [{\"name\": \"regcred\"}]}' -n memphis-$unique_id"
      //sh "sleep 40"
    }

    stage('Build and push docker image to Docker Hub') {
       sh "docker buildx build --push -t ${repoUrlPrefix}/${imageName}-${gitBranch}-${test_suffix} ."
    }

    stage('Tests - Install/upgrade Memphis cli') {
      sh "sudo npm uninstall memphis-dev-cli"
      sh "sudo npm i memphis-dev-cli -g"
    }

    ////////////////////////////////////////
    //////////// Docker-Compose ////////////
    ////////////////////////////////////////

    stage('Tests - Docker compose install') {
      sh "rm -rf memphis-docker"
      dir ('memphis-docker'){
        git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-docker.git', branch: gitBranch
      }
        sh "docker-compose -f ./memphis-docker/docker-compose-dev-tests-broker.yml -p memphis up -d"
    }

    stage('Tests - Run e2e tests over Docker') {
      sh "rm -rf memphis-e2e-tests"
      dir ('memphis-e2e-tests'){
        git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-e2e-tests.git', branch: 'master'
      }
      sh "npm install --prefix ./memphis-e2e-tests"
      sh "node ./memphis-e2e-tests/index.js docker"
    }

    stage('Tests - Remove Docker compose') {
      sh "docker-compose -f ./memphis-docker/docker-compose-dev-tests-broker.yml -p memphis down"
      sh "docker volume prune -f"
    }

    ////////////////////////////////////////
    ////////////   Kubernetes   ////////////
    ////////////////////////////////////////

    stage('Tests - Install memphis with helm') {
      	sh "rm -rf memphis-k8s"
      	dir ('memphis-k8s'){
       	    git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-k8s.git', branch: gitBranch
            sh "helm upgrade --atomic --install memphis-tests memphis --set analytics='false',teston='cp' --create-namespace --namespace memphis-$unique_id"
      	}
    }


    stage('Open port forwarding to memphis service') {
      sh(script: """until kubectl get pods --selector=app.kubernetes.io/name=memphis -o=jsonpath="{.items[*].status.phase}" -n memphis-$unique_id  | grep -q "Running" ; do sleep 1; done""", returnStdout: true)
      sh "nohup kubectl port-forward service/memphis-cluster 6666:6666 9000:9000 --namespace memphis-$unique_id &"
    }

    stage('Tests - Run e2e tests over kubernetes') {
      sh "npm install --prefix ./memphis-e2e-tests"
      sh "node ./memphis-e2e-tests/index.js kubernetes memphis-$unique_id"
    }

    stage('Tests - Uninstall helm') {
      sh "helm uninstall memphis-tests -n memphis-$unique_id"
      sh "kubectl delete ns memphis-$unique_id &"
      sh(script: """/usr/sbin/lsof -i :6666,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh""", returnStdout: true)
    }


    ////////////////////////////////////////
    ////////////  Build & Push  ////////////
    ////////////////////////////////////////


    stage('Build and push image to Docker Hub') {
      sh "docker buildx use builder"
      if (env.BRANCH_NAME ==~ /(master)/) { //NEW TAG
	sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}-${gitBranch} --platform linux/amd64,linux/arm64 ."
      }
      else{
	sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}:${versionTag} --tag ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 ."	
      }
    }


    //////////////////////////////////////
    //////////////  MASTER  //////////////
    //////////////////////////////////////

      if (env.BRANCH_NAME ==~ /(master)/) {
    	stage('Push to staging'){
	  sh "aws eks --region eu-central-1 update-kubeconfig --name staging-cluster"
          sh "helm uninstall my-memphis --kubeconfig ~/.kube/config -n memphis"
	  sh(script: """kubectl get pvc -n memphis | grep -v NAME| awk '{print\$1}' | while read vol; do kubectl delete pvc \$vol -n memphis; done""", returnStdout: true )
          sh 'helm install --wait my-memphis memphis-k8s/memphis --set analytics="false",cluster.enabled="true" --kubeconfig ~/.kube/config --create-namespace --namespace memphis'
          sh "rm -rf memphis-k8s"
	}
	      
	stage('Open port forwarding to memphis service') {
          sh(script: """until kubectl get pods --selector=app.kubernetes.io/name=memphis -o=jsonpath="{.items[*].status.phase}" -n memphis  | grep -q "Running" ; do sleep 1; done""", returnStdout: true)
     	  sh "nohup kubectl port-forward service/memphis-cluster 6666:6666 9000:9000 --namespace memphis &"
   	}

   	stage('Tests - Run e2e tests over memphis cluster') {
          sh "npm install --prefix ./memphis-e2e-tests"
          sh "node ./memphis-e2e-tests/index.js kubernetes memphis"
        }

    	stage('Tests - remove port-forwarding') {
          sh(script: """/usr/sbin/lsof -i :6666,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh""", returnStdout: true)
        }

        stage('Tests - Remove used directories') {
       	  sh "rm -rf memphis-e2e-tests"
    	}
      }

    //////////////////////////////////////////////////////////
    //////////////  Checkout to version branch  //////////////
    //////////////////////////////////////////////////////////

      if (env.BRANCH_NAME ==~ /(latest)/) {
    	stage('checkout to version branch'){
	    withCredentials([sshUserPrivateKey(keyFileVariable:'check',credentialsId: 'main-github')]) {
	    sh "git reset --hard origin/latest"
	    sh "GIT_SSH_COMMAND='ssh -i $check'  git checkout -b ${versionTag}"
       	    sh "GIT_SSH_COMMAND='ssh -i $check' git push --set-upstream origin ${versionTag}"
  	  }
	}
      }  

    notifySuccessful()
  } catch (e) {
      currentBuild.result = "FAILED"
      sh(script: """docker ps | grep memphisos/ | awk '{print \"docker rm -f \"\$1}' | sh""", returnStdout: true)
      sh "docker volume prune -f"
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
      recipientProviders: [requestor()]
    )
}
def notifyFailed() {
  emailext (
      subject: "FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
      body: """<p>FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':</p>
        <p>Check console output at &QUOT;<a href='${env.BUILD_URL}'>${env.JOB_NAME} [${env.BUILD_NUMBER}]</a>&QUOT;</p>""",
      recipientProviders: [requestor()]
    )
}
