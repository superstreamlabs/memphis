String unique_id = org.apache.commons.lang.RandomStringUtils.random(4, false, true)
pipeline {
  environment {
      versionTag= readFile('./version.conf')
      gitBranch = "${env.BRANCH_NAME}"
      imageName = "memphis"
      repoUrlPrefix = "memphisos"
      test_suffix = "test"
  }

  agent {
    label 'memphis-jenkins-big-fleet,'
  }

  stages {
    stage('Login to Docker Hub') {
        steps {
            withCredentials([usernamePassword(credentialsId: 'docker-hub', usernameVariable: 'DOCKER_HUB_CREDS_USR', passwordVariable: 'DOCKER_HUB_CREDS_PSW')]) {
                sh 'docker login -u $DOCKER_HUB_CREDS_USR -p $DOCKER_HUB_CREDS_PSW'
            }
        }
    }
	  
	
    stage('Create memphis namespace in Kubernetes'){
        steps {
            sh """
	    	minikube start
                kubectl config use-context minikube
                kubectl create namespace memphis-$unique_id --dry-run=client -o yaml | kubectl apply -f -
                aws s3 cp s3://memphis-jenkins-backup-bucket/regcred.yaml .
                kubectl apply -f regcred.yaml -n memphis-$unique_id
                kubectl patch serviceaccount default -p '{\"imagePullSecrets\": [{\"name\": \"regcred\"}]}' -n memphis-$unique_id
            """
        }
    }

    stage('Build and push docker image to Docker Hub') {
        steps {
            sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}-${gitBranch} --platform linux/amd64,linux/arm64 ."
        }
    }

    stage('Tests - Install/upgrade Memphis cli - BETA') {
        when { anyOf { branch 'master'; branch 'qa'}}
        steps {
            sh """
            sudo npm uninstall memphis-dev-cli-beta -g
            sudo npm i memphis-dev-cli-beta -g --force
            """
        }
    }

    stage('Tests - Install/upgrade Memphis cli - LATEST') {
        when { branch 'latest' }
        steps {
            sh """
            sudo npm uninstall memphis-dev-cli -g
            sudo npm i memphis-dev-cli -g
            """
        }
    }



    ////////////////////////////////////////
    //////////// Docker-Compose ////////////
    ////////////////////////////////////////

    stage('Tests - Docker compose install - Master') {
        when { branch 'master' }
        steps {
            sh "rm -rf memphis-docker"
            dir ('memphis-devops'){
                git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-devops.git', branch: 'master'
            }
            sh "docker-compose -f ./memphis-devops/docker/docker-compose-master-tests-broker.yml -p memphis up -d"
        }
    }
    stage('Tests - Docker compose install - QA') {
        when { branch 'qa' }
        steps {
            sh "rm -rf memphis-docker"
            dir ('memphis-devops'){
                git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-devops.git', branch: 'master'
            }
            sh "docker-compose -f ./memphis-devops/docker/docker-compose-qa-tests-broker.yml -p memphis up -d"
        }
    }
	  
    stage('Tests - Docker compose install - Latest') {
        when { branch 'latest' }
        steps {
            sh "rm -rf memphis-docker"
            dir ('memphis-devops'){
                git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-devops.git', branch: 'master'
            }
            sh "docker-compose -f ./memphis-devops/docker/docker-compose-latest-tests-broker.yml -p memphis up -d"
        }
    }


    stage('Tests - Run e2e tests over Docker') {
        steps {
            sh "rm -rf memphis-e2e-tests"
            dir ('memphis-e2e-tests'){
                git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-e2e-tests.git', branch: 'master'
           }
            sh "npm install --prefix ./memphis-e2e-tests"
            sh "node ./memphis-e2e-tests/index.js docker"
        }
    }

    stage('Tests - Remove Docker compose - Master') {
        when { branch 'master' }
        steps {
            sh """
            docker-compose -f ./memphis-devops/docker/docker-compose-master-tests-broker.yml -p memphis down
            docker volume prune -f
            """
        }
    }
    stage('Tests - Remove Docker compose - QA') {
        when { branch 'qa' }
        steps {
            sh """
            docker-compose -f ./memphis-devops/docker/docker-compose-qa-tests-broker.yml -p memphis down
            docker volume prune -f
            """
        }
    }

    stage('Tests - Remove Docker compose - Latest') {
        when { branch 'latest' }
        steps {
            sh """
            docker-compose -f ./memphis-devops/docker/docker-compose-latest-tests-broker.yml -p memphis down
            docker volume prune -f
            """
        }
    }

    ////////////////////////////////////////
    ////////////   Kubernetes   ////////////
    ////////////////////////////////////////

    stage('Tests - Install memphis with helm') {
        steps {
      	    dir ('memphis-k8s'){
       	        git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-k8s.git', branch: 'master'
                sh """
	                helm install memphis-tests memphis --set memphis.extraEnvironmentVars.enabled=true,memphis.image=${repoUrlPrefix}/${imageName}-${gitBranch} --set-json 'memphis.extraEnvironmentVars.vars=[{"name":"ENV","value":"staging"}]' --create-namespace --namespace memphis-$unique_id --wait
                """
      	    }
        }
    }


    stage('Open port forwarding to memphis service - Minikube') {
        steps {
            sh """
                until kubectl get pods --selector=app.kubernetes.io/name=memphis -o=jsonpath="{.items[*].status.containerStatuses[*].ready}" -n memphis-$unique_id  | grep -v "false" ; do sleep 1; done
                nohup kubectl port-forward service/memphis 6666:6666 9000:9000 7770:7770 --namespace memphis-$unique_id &
            """
        }
    }

    stage('Tests - Run e2e tests over kubernetes - Minikube') {
        steps {
            sh """
                npm install --prefix ./memphis-e2e-tests
                node ./memphis-e2e-tests/index.js kubernetes memphis-$unique_id
            """
        }
    }

    stage('Tests - Uninstall helm') {
        steps {
            sh """
                helm uninstall memphis-tests -n memphis-$unique_id
                kubectl delete ns memphis-$unique_id
                lsof -i :6666,9000,7770 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh
            """
        }
    }


    ////////////////////////////////////////
    ////////////  Build & Push  ////////////
    ////////////////////////////////////////
		
    stage('Build and push image to Docker Hub - LATEST') {
		when { branch 'latest' }
		steps {	
        	sh """
                docker buildx build --push --tag ${repoUrlPrefix}/${imageName}:${versionTag} --tag ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 .
            """
        }
    }

    //////////////////////////////////////
    ////////////  K8's Tests  ////////////
    //////////////////////////////////////

    stage('Reset STG-OSS environment') {
        when { not {branch 'latest'}}
        steps {
            sh """gcloud container clusters get-credentials memphis-staging-gke --region europe-west3 --project memphis-k8s-staging"""
	    catchError(buildResult: 'SUCCESS', message: 'helm uninstall failed because memphis was not deployed to this namespace') {
	      sh """helm uninstall my-memphis --kubeconfig ~/.kube/config -n memphis"""
            }
	    sh """kubectl get pvc -n memphis | grep -v NAME| awk '{print\$1}' | while read vol; do kubectl delete pvc \$vol -n memphis; done"""
        }
    }

    stage('Push to STG-OSS') {
        when { not {branch 'latest'}}
        steps {
            dir ('memphis-k8s'){
       	        git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-k8s.git', branch: 'master'
	            sh """
                    gsutil cp gs://memphis-jenkins-backup-bucket/memphis-staging-oss.yaml .
                    helm install my-memphis memphis --set memphis.image=${repoUrlPrefix}/${imageName}-${gitBranch} -f ./memphis-staging-oss.yaml --create-namespace --namespace memphis --wait
	            """
	        }
          sh "rm -rf memphis-k8s"
        }
    }

    stage('Open port forwarding to memphis service - K8s') {
        when { not {branch 'latest'}}
        steps {
            sh """
                until kubectl get pods --selector=app.kubernetes.io/name=memphis -o=jsonpath="{.items[*].status.containerStatuses[*].ready}" -n memphis  | grep -v "false" ; do sleep 1; done
                nohup kubectl port-forward service/memphis 6666:6666 9000:9000 7770:7770 --namespace memphis &
            """
        }
    }

    stage('Tests - Run e2e tests over kubernetes - K8s') {
        when { not {branch 'latest'}}
        steps {
            sh """
                npm install --prefix ./memphis-e2e-tests
                node ./memphis-e2e-tests/index.js kubernetes memphis
            """
        }
    }

    stage('Install memphis CLI') {
        when { not {branch 'latest'}}
        steps {
            sh """
                sudo npm i memphis-dev-cli -g
            """
        }
    }
	      
   	stage('Create staging user') {
        when { not {branch 'latest'}}
        steps {
   	        withCredentials([string(credentialsId: 'staging_pass', variable: 'staging_pass')]) {
   	        sh """
     	        mem connect -s localhost -u root -p \$(kubectl get secret memphis-creds  -n memphis -o jsonpath="{.data.ROOT_PASSWORD}" | base64 --decode)
     	        mem user add -u staging -p $staging_pass
    	        lsof -i :6666,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh
    	    """
            }
   	    }
   	}
	      
    stage('Tests - remove port-forwarding') {
        when { not {branch 'latest'}}
        steps {
            sh"""
                /usr/sbin/lsof -i :6666,9000 | grep kubectl | awk '{print \"kill -9 \"\$2}' | sh
            """
        }
    }

    stage('Tests - Remove used directories') {
        when { not {branch 'latest'}}
       	steps {
            sh "rm -rf memphis-e2e-tests"
        }
    }

    //////////////////////////////////////////////////////////
    //////////////  Checkout to version branch  //////////////
    //////////////////////////////////////////////////////////

    stage('checkout to version branch'){
        when { branch 'latest' }
        steps {
	        withCredentials([sshUserPrivateKey(keyFileVariable:'check',credentialsId: 'main-github')]) {
	            sh """
	                git reset --hard origin/latest
	                GIT_SSH_COMMAND='ssh -i $check'  git checkout -b ${versionTag}
       	            GIT_SSH_COMMAND='ssh -i $check' git push --set-upstream origin ${versionTag}
	            """
            }
        }
	}

    stage('Install gh + create new release'){
        when { branch 'latest' }
        steps {
	        withCredentials([sshUserPrivateKey(keyFileVariable:'check',credentialsId: 'main-github')]) {
	            sh """
	                sudo dnf config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo -y
                    sudo dnf install gh -y
                    gh release create v${versionTag} --generate-notes
	            """
            }
        }
    }
  }
    post {
        always {
            cleanWs()
        }
        success {
            notifySuccessful()
        }

        failure {
            notifyFailed()
        }
    }
}
def notifySuccessful() {
    emailext (
        subject: "SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
        body: """SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':
        Check console output and connection attributes at ${env.BUILD_URL}""",
        recipientProviders: [requestor()]
    )
}
def notifyFailed() {
    emailext (
        subject: "FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'",
        body: """FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]':
        Check console output at ${env.BUILD_URL}""",
        recipientProviders: [requestor()]
    )
}
