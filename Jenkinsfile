def gitBranch = env.BRANCH_NAME
def imageName = "memphis"
def gitURL = "git@github.com:Memphisdev/memphis.git"
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
/* 
    stage('UI build'){
      dir ('ui_src'){
	sh """
	  npm install
	  CI=false npm run build
	"""
      }
    }
*/	  
    stage('Create memphis namespace in Kubernetes'){
      sh """
        kubectl config use-context minikube
        kubectl create namespace memphis-$unique_id --dry-run=client -o yaml | kubectl apply -f -
        aws s3 cp s3://memphis-jenkins-backup-bucket/regcred.yaml .
        kubectl apply -f regcred.yaml -n memphis-$unique_id
        kubectl patch serviceaccount default -p '{\"imagePullSecrets\": [{\"name\": \"regcred\"}]}' -n memphis-$unique_id
      """
    }

    stage('Build and push docker image to Docker Hub') {
       sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}-${gitBranch} --platform linux/amd64,linux/arm64 ."
    }

    stage('Tests - Install/upgrade Memphis cli') {
      if (env.BRANCH_NAME ==~ /(master)/) { 
        sh """
          sudo npm uninstall memphis-dev-cli-beta -g
          sudo npm i memphis-dev-cli-beta -g --force
        """
      }
      else {
	sh """
          sudo npm uninstall memphis-dev-cli -g
          sudo npm i memphis-dev-cli -g
        """
      }
    }

    ////////////////////////////////////////
    //////////// Docker-Compose ////////////
    ////////////////////////////////////////

    stage('Tests - Docker compose install') {
      sh "rm -rf memphis-docker"
      dir ('memphis-devops'){
        git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-devops.git', branch: gitBranch
      }
			if (env.BRANCH_NAME ==~ /(latest)/) {
        sh "docker-compose -f ./memphis-docker/docker-compose-latest-tests-broker.yml -p memphis up -d"
			}
			if (env.BRANCH_NAME ==~ /(master)/) {
				sh "docker-compose -f ./memphis-docker/docker-compose-master-tests-broker.yml -p memphis up -d"
			}
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
			if (env.BRANCH_NAME ==~ /(latest)/) {
        sh """
          docker-compose -f ./memphis-docker/docker-compose-latest-tests-broker.yml -p memphis down
          docker volume prune -f
        """
			}
			if (env.BRANCH_NAME ==~ /(master)/) {
				sh """
          docker-compose -f ./memphis-docker/docker-compose-master-tests-broker.yml -p memphis down
          docker volume prune -f
        """
			}
    }

    ////////////////////////////////////////
    ////////////   Kubernetes   ////////////
    ////////////////////////////////////////

    stage('Tests - Install memphis with helm') {
      	sh "rm -rf memphis-k8s"
      	dir ('memphis-k8s'){
       	  git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-k8s.git', branch: gitBranch
          sh """
	    helm install memphis-tests memphis --set memphis.extraEnvironmentVars.enabled=true,memphis.image=${repoUrlPrefix}/${imageName}-${gitBranch} --set-json 'memphis.extraEnvironmentVars.vars=[{"name":"ENV","value":"staging"}]' --create-namespace --namespace memphis-$unique_id --wait
          """
      	}
    }


    stage('Open port forwarding to memphis service') {
      sh """
        until kubectl get pods --selector=app.kubernetes.io/name=memphis -o=jsonpath="{.items[*].status.phase}" -n memphis-$unique_id  | grep -q "Running" ; do sleep 1; done
        nohup kubectl port-forward service/memphis 6666:6666 9000:9000 7770:7770 --namespace memphis-$unique_id &
      """
    }

    stage('Tests - Run e2e tests over kubernetes') {
      sh """
        npm install --prefix ./memphis-e2e-tests
        node ./memphis-e2e-tests/index.js kubernetes memphis-$unique_id
      """
    }

    stage('Tests - Uninstall helm') {
      sh """
        helm uninstall memphis-tests -n memphis-$unique_id
        kubectl delete ns memphis-$unique_id &
        /usr/sbin/lsof -i :6666,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh
      """
    }


    ////////////////////////////////////////
    ////////////  Build & Push  ////////////
    ////////////////////////////////////////


    stage('Build and push image to Docker Hub') {
      sh "docker buildx use builder"
      if (env.BRANCH_NAME ==~ /(latest)/) {
	if(versionTag.contains('stable')) {
	  sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}:${versionTag} --tag ${repoUrlPrefix}/${imageName}:stable --tag ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 ."	
	}
	else{
          sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}:${versionTag} --tag ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 ."	
	}      
      }
    }


    //////////////////////////////////////
    //////////////  MASTER  //////////////
    //////////////////////////////////////

      if (env.BRANCH_NAME ==~ /(master)/) {
    	stage('Push to staging'){
	  sh """
	    aws eks --region eu-central-1 update-kubeconfig --name staging-cluster
            helm uninstall my-memphis --kubeconfig ~/.kube/config -n memphis
	    kubectl get pvc -n memphis | grep -v NAME| awk '{print\$1}' | while read vol; do kubectl delete pvc \$vol -n memphis; done
	  """
	  dir ('memphis-k8s'){
       	    git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-k8s.git', branch: gitBranch
	    sh """
              aws s3 cp s3://memphis-jenkins-backup-bucket/memphis-staging-oss.yaml .
              helm install my-memphis memphis --set global.cluster.enabled="true",websocket.tls.cert="tls.crt",websocket.tls.key="tls.key",websocket.tls.secret.name="ws-tls-certs" -f ./memphis-staging-oss.yaml --create-namespace --namespace memphis --wait
	    """
	  }
          sh "rm -rf memphis-k8s"
	}
	      
	stage('Open port forwarding to memphis service') {
          sh """
	    until kubectl get pods --selector=app.kubernetes.io/name=memphis -o=jsonpath="{.items[*].status.phase}" -n memphis  | grep -q "Running" ; do sleep 1; done
     	    nohup kubectl port-forward service/memphis 6666:6666 9000:9000 7770:7770 --namespace memphis &
	  """
   	}

   	stage('Tests - Run e2e tests over memphis cluster') {
          sh """
	    npm install --prefix ./memphis-e2e-tests
            node ./memphis-e2e-tests/index.js kubernetes memphis
	  """
        }

	stage('Install memphis CLI') {
        sh """
          sudo npm i memphis-dev-cli -g
        """
        }
	      
   	stage('Create staging user') {
   	  withCredentials([string(credentialsId: 'staging_pass', variable: 'staging_pass')]) {
   	    sh '''
     	    mem connect -s localhost -u root -p \$(kubectl get secret memphis-creds  -n memphis -o jsonpath="{.data.ROOT_PASSWORD}" | base64 --decode)
     	    mem user add -u staging -p $staging_pass
    	    /usr/sbin/lsof -i :6666,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh
    	  '''
   	  }
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
	      sh """
	        git reset --hard origin/latest
	        GIT_SSH_COMMAND='ssh -i $check'  git checkout -b ${versionTag}
       	        GIT_SSH_COMMAND='ssh -i $check' git push --set-upstream origin ${versionTag}
	      """
  	  }
	}
	      
	stage('Install gh'){
	  sh """
	    sudo dnf config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo -y
            sudo dnf install gh -y
	  """
	}
	      
	stage('Create new release') {
          withCredentials([string(credentialsId: 'gh_token', variable: 'GH_TOKEN')]) {
	    sh "gh release create v${versionTag} --generate-notes"
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
