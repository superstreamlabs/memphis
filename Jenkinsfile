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


    stage('Build and push image to Docker Hub') {
    	  sh "docker buildx build --push --tag ${repoUrlPrefix}/${imageName}-${gitBranch}:${versionTag} --tag ${repoUrlPrefix}/${imageName}-${gitBranch} --platform linux/amd64,linux/arm64 ."
    }
	
    ///////////////////////////////////////
    //////////////  SANDBOX  //////////////
    ///////////////////////////////////////
    stage('Remove memphis'){
			sh "aws eks --region eu-central-1 update-kubeconfig --name sandbox-cluster"
			sh "helm uninstall my-memphis -n memphis"
			sh "kubectl delete ns memphis"
		}
    stage('Create memphis namespace in Kubernetes'){
      sh "kubectl create namespace memphis --dry-run=client -o yaml | kubectl apply -f -"
      sh "aws s3 cp s3://memphis-jenkins-backup-bucket/regcred.yaml ."
      sh "kubectl apply -f regcred.yaml -n memphis"
      sh "kubectl patch serviceaccount default -p '{\"imagePullSecrets\": [{\"name\": \"regcred\"}]}' -n memphis"
    }
	  
      stage('Push to sandbox'){
				sh "rm -rf memphis-infra"
      	dir ('memphis-infra'){
       	  git credentialsId: 'main-github', url: 'git@github.com:memphisdev/memphis-infra.git', branch: gitBranch
      	}
      	sh "helm upgrade --atomic --install my-memphis memphis-infra/kubernetes/helm/memphis --set cluster.enabled='true',analytics='false',sandbox='true' --create-namespace --namespace memphis"
				sh "rm -rf memphis-infra"
      }
      
	  
      /*stage('Configure sandbox URLs'){
				sh "aws s3 cp s3://memphis-jenkins-backup-bucket/sandbox_files/update_sandbox_record.json ." //sandbox.memphis.dev redirect to new LB record
				sh(script: """sed "s/\\"DNSName\\": \\"\\"/\\"DNSName\\": \\"\$(aws elbv2 describe-load-balancers --load-balancer-arns | grep "k8s-memphis-memphisu" | grep DNS | awk '{print \"dualstack.\"\$2}' | sed 's/"//g' | sed 's/,//g')\\"/g"  update_sandbox_record.json > record1.json""", returnStdout: true)    
				sh(script: """aws route53 change-resource-record-sets --hosted-zone-id Z05132833CK9UXS6W3I0E --change-batch file://record1.json > status1.txt""",    returnStdout: true)
	
				//broker url section      
				sh "aws s3 cp s3://memphis-jenkins-backup-bucket/sandbox_files/update_broker_record.json ."  //broker.sandbox.memphis.dev redirect to new LB record
				sh(script: """sed "s/\\"DNSName\\": \\"\\"/\\"DNSName\\": \\"\$(kubectl get svc -n memphis | grep "memphis-cluster-sandbox" | awk '{print \"dualstack.\"\$4}')\\"/g"  update_broker_record.json > record2.json""",  returnStdout: true)
				sh(script: """aws route53 change-resource-record-sets --hosted-zone-id Z05132833CK9UXS6W3I0E --change-batch file://record2.json > status2.txt""",    returnStdout: true) 
				sh "rm -rf record1.json record2.json update_sandbox_record.json update_broker_record.json"
      }*/
  
	  
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
