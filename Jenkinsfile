def dockerImagesRepo = "strechinc"
def imageName = "strech-server-beta"
def gitURL = "git@github.com:STRECH-LTD/strech-server.git"
def gitBranch = "beta"
unique_Id = UUID.randomUUID().toString()
def DOCKER_HUB_CREDS = credentials('docker-hub')

node {
  try{
    stage('SCM checkout') {
        git credentialsId: 'main-github', url: gitURL, branch: gitBranch
    }
    stage('Build docker image') {
        sh "docker build -t ${dockerImagesRepo}/${imageName} ."
    }

    stage('Push docker image') {
	withCredentials([usernamePassword(credentialsId: 'docker-hub', usernameVariable: 'DOCKER_HUB_CREDS_USR', passwordVariable: 'DOCKER_HUB_CREDS_PSW')]) {
		sh "docker login -u $DOCKER_HUB_CREDS_USR -p $DOCKER_HUB_CREDS_PSW"
	        sh "docker tag ${dockerImagesRepo}/${imageName} ${dockerImagesRepo}/${imageName}:${unique_Id}"
		sh "docker push ${dockerImagesRepo}/${imageName}:${unique_Id}"
		sh "docker push ${dockerImagesRepo}/${imageName}:latest"
		sh "docker image rm ${dockerImagesRepo}/${imageName}:latest"
		sh "docker image rm ${dockerImagesRepo}/${imageName}:${unique_Id}"
	}
    }
    
    //stage('Push image to kubernetes') {
//	    sh "kubectl --kubeconfig=\"/var/lib/jenkins/.kube/strech-staging-kubeconfig.yaml\" apply -f \"Staging/k8s-template.yaml\" --record"
  //    sh "kubectl --kubeconfig=\"/var/lib/jenkins/.kube/strech-staging-kubeconfig.yaml\" set image deployment/${imageName} ${imageName}=${dockerImagesRepo}/${imageName}:${unique_Id} -n hub"
    //}
    notifySuccessful()

  } catch (e) {
      currentBuild.result = "FAILED"
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
