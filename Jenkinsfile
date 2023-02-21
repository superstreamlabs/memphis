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
	  

    ////////////////////////////////////////
    ////////////  Build & Push  ////////////
    ////////////////////////////////////////
    stage('UI build') {
      dir ('ui_src'){
	sh """
	  npm install
	  CI=false npm run build
	"""
      }
    }

    stage('Build and push image to Docker Hub') {
      sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}-${gitBranch}:${versionTag} --tag ${repoUrlPrefix}/${imageName}-${gitBranch} --platform linux/amd64,linux/arm64 ."
    }
	
    ///////////////////////////////////////
    //////////////  SANDBOX  //////////////
    ///////////////////////////////////////
    stage('Remove memphis'){
      sh """
        aws eks --region eu-central-1 update-kubeconfig --name sandbox-cluster
        helm uninstall my-memphis -n memphis
        kubectl delete ns memphis
      """
   }
	  
    stage('Create memphis namespace in Kubernetes'){
      sh """
        kubectl create namespace memphis --dry-run=client -o yaml | kubectl apply -f -
        aws s3 cp s3://memphis-jenkins-backup-bucket/regcred.yaml .
        kubectl apply -f regcred.yaml -n memphis
        kubectl patch serviceaccount default -p '{\"imagePullSecrets\": [{\"name\": \"regcred\"}]}' -n memphis
      """
    }
     
    stage('Create new secret in memphis namespace'){
      withCredentials([file(credentialsId: 'memphis.pem', variable: 'cert'), file(credentialsId: 'memphis-key.pem', variable: 'key')]) {
        sh "kubectl create secret generic tls-secret --from-file=$cert --from-file=$key -n memphis"
      }
    }
	  
      stage('Push to sandbox'){
        sh "rm -rf memphis-sbox-k8s"
      	dir ('memphis-sbox-k8s'){
       	  git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-sbox-k8s.git', branch: gitBranch
	  sh "helm install my-memphis memphis --set cluster.enabled='true',analytics='false',sandbox='true' --create-namespace --namespace memphis --wait"
      	}
   	sh "rm -rf memphis-sbox-k8s"
      }
      
	  
      stage('Configure sandbox URLs'){
        //UI url section
	sh "aws s3 cp s3://memphis-jenkins-backup-bucket/sandbox_files/update_sandbox_record.json ." //sandbox.memphis.dev redirect to new LB record
	sh(script: """sed "s/\\"DNSName\\": \\"\\"/\\"DNSName\\": \\"\$(kubectl get svc -n memphis | grep "memphis-cluster-sandbox" | awk '{print \"dualstack.\"\$4}')\\"/g"  update_sandbox_record.json > record1.json""", returnStdout: true)    
	sh(script: """aws route53 change-resource-record-sets --hosted-zone-id Z05132833CK9UXS6W3I0E --change-batch file://record1.json > status1.txt""",    returnStdout: true)
	
	//broker url section      
	sh "aws s3 cp s3://memphis-jenkins-backup-bucket/sandbox_files/update_broker_record.json ."  //broker.sandbox.memphis.dev redirect to new LB record
	sh(script: """sed "s/\\"DNSName\\": \\"\\"/\\"DNSName\\": \\"\$(kubectl get svc -n memphis | grep "memphis-cluster-sandbox" | awk '{print \"dualstack.\"\$4}')\\"/g"  update_broker_record.json > record2.json""",  returnStdout: true)
	sh(script: """aws route53 change-resource-record-sets --hosted-zone-id Z05132833CK9UXS6W3I0E --change-batch file://record2.json > status2.txt""",    returnStdout: true) 

	//proxy url section      
	sh "aws s3 cp s3://memphis-jenkins-backup-bucket/sandbox_files/update_restgw_record.json ."  //restgw.sandbox.memphis.dev redirect to new LB record
	sh(script: """sed "s/\\"DNSName\\": \\"\\"/\\"DNSName\\": \\"\$(kubectl get svc -n memphis | grep "memphis-rest-gateway" | awk '{print \"dualstack.\"\$4}')\\"/g"  update_proxy_record.json > record3.json""",  returnStdout: true)
	sh(script: """aws route53 change-resource-record-sets --hosted-zone-id Z05132833CK9UXS6W3I0E --change-batch file://record3.json > status3.txt""",    returnStdout: true)       
	sh "rm -rf record1.json record2.json record3.json update_sandbox_record.json update_broker_record.json update_restgw_record.json"
      }
  
      stage('Run memphis-demo CI/CD'){
        build job: '../../Memphis-Sidecars/memphis-demo'
      }
	  
    notifySuccessful()
	  
  } catch (e) {
      currentBuild.result = "FAILED"
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
