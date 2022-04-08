def imageName = "memphis-server-staging"
def containerName = "memphis-server"
def gitURL = "git@github.com:Memphis-OS/memphis-server.git"
def gitBranch = "staging"
def repoUrlPrefix = "221323242847.dkr.ecr.eu-central-1.amazonaws.com"
unique_Id = UUID.randomUUID().toString()
def namespace = "memphis"

node {
  try{
    stage('SCM checkout') {
        git credentialsId: 'main-github', url: gitURL, branch: gitBranch
    }
    stage('Build docker image') {
        sh "docker build -t ${repoUrlPrefix}/${imageName} ."
    }

    stage('Push docker image') {
	sh "aws ecr describe-repositories --repository-names ${imageName} --region eu-central-1 || aws ecr create-repository --repository-name ${imageName} --region eu-central-1 && aws ecr put-lifecycle-policy --repository-name ${imageName} --region eu-central-1 --lifecycle-policy-text 'file:///var/lib/jenkins/utils/ecr-lifecycle-policy.json'"
        sh "docker tag ${repoUrlPrefix}/${imageName} ${repoUrlPrefix}/${imageName}:${unique_Id}"
        sh "aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin 221323242847.dkr.ecr.eu-central-1.amazonaws.com"
        sh "docker push ${repoUrlPrefix}/${imageName}:${unique_Id}"
        sh "docker push ${repoUrlPrefix}/${imageName}:latest"
        sh "docker image rm ${repoUrlPrefix}/${imageName}:latest"
        sh "docker image rm ${repoUrlPrefix}/${imageName}:${unique_Id}"
    }
    
    stage('Push image to kubernetes') {
	sh "kubectl --kubeconfig=\"/var/lib/jenkins/.kube/memphis-staging-kubeconfig.yaml\" apply -f k8s-template.yaml --record -n ${namespace}"
	sh "kubectl --kubeconfig=\"/var/lib/jenkins/.kube/memphis-staging-kubeconfig.yaml\" set image deployment/${containerName} ${containerName}=${repoUrlPrefix}/${imageName}:${unique_Id} -n ${namespace}"
    }
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
