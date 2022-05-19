def imageName = "memphis-control-plane-staging"
def containerName = "memphis-control-plane"
def gitURL = "git@github.com:Memphis-OS/memphis-control-plane.git"
def gitBranch = "staging"
def repoUrlPrefix = "221323242847.dkr.ecr.eu-central-1.amazonaws.com"
unique_Id = UUID.randomUUID().toString()
def namespace = "memphis"
def test_suffix = "test"

node {
  try{
    stage('SCM checkout') {
        git credentialsId: 'main-github', url: gitURL, branch: gitBranch
    }
    stage('Login to ECR') {
      sh "aws ecr get-login-password --region eu-central-1 | docker login --username AWS --password-stdin ${repoUrlPrefix}"
    }

    stage('Create ECR Repo') {
      sh "aws ecr describe-repositories --repository-names ${imageName}-${test_suffix} --region eu-central-1 || aws ecr create-repository --repository-name ${imageName}-${test_suffix} --region eu-central-1 && aws ecr put-lifecycle-policy --repository-name ${imageName}-${test_suffix} --region eu-central-1 --lifecycle-policy-text 'file:///var/lib/jenkins/utils/ecr-lifecycle-policy.json'"
    }
	  
    stage('Build and push docker image to ECR') {
      sh "docker buildx build --push -t ${repoUrlPrefix}/${imageName}-${test_suffix} --platform linux/amd64,linux/arm64 ."
    }

    stage('Tests - Install/upgrade Memphis cli') {
      sh "sudo npm uninstall memphis-dev-cli"
      sh "sudo npm i memphis-dev-cli -g"
    }

    stage('Tests - Docker compose install') {
      sh "docker-compose -f /var/lib/jenkins/tests/docker-compose-files/docker-compose-dev-memphis-control-plane.yml -p memphis up -d"
    }

    stage('Tests - Run e2e tests over docker') {
      sh "git clone git@github.com:Memphis-OS/memphis-k8s.git"
      sh "cd memphis-e2e-tests"
      sh "npm install"
      sh "node index.js docker"
      sh "cd ../"
    }

    stage('Tests - Remove Docker compose') {
      sh "docker-compose -f /var/lib/jenkins/tests/docker-compose-files/docker-compose-dev-memphis-control-plane.yml -p memphis down"
    }

    stage('Tests - Install helm') {
      sh "rm -rf memphis-k8s"
      sh "git clone --branch tests git@github.com:Memphis-OS/memphis-k8s.git"
      sh 'helm install memphis-tests memphis-k8s/helm/memphis --set analytics="false",test-on="cp" --create-namespace --namespace memphis'
    }

    stage('Tests - Run e2e tests over helm/k8s') {
      sh "cd memphis-e2e-tests"
      sh "npm install"
      sh "node index.js k8s"
      sh "cd ../"
    }

    stage('Tests - Uninstall helm') {
      sh "helm uninstall memphis -n memphis"
      sh "kubectl delete ns memphis"
    }

    stage('Tests - Remove e2e-tests') {
      sh "rm -rf memphis-e2e-tests"
    }

    stage('Delete ECR Test repo'){
      sh "aws ecr delete-repository --repository-name ${imageName}-${test_suffix} --region eu-central-1"
    }

    stage('Create Staging ECR Repo') {
      sh "aws ecr describe-repositories --repository-names ${imageName} --region eu-central-1 || aws ecr create-repository --repository-name ${imageName} --region eu-central-1 && aws ecr put-lifecycle-policy --repository-name ${imageName} --region eu-central-1 --lifecycle-policy-text 'file:///var/lib/jenkins/utils/ecr-lifecycle-policy.json'"
    }
	  
    stage('Build and push docker image to ECR') {
      sh "docker buildx build --push -t ${repoUrlPrefix}/${imageName} --platform linux/amd64,linux/arm64 ."
    }

    stage('Push to staging'){
      sh "helm uninstall my-memphis --kubeconfig /var/lib/jenkins/.kube/memphis-staging-kubeconfig.yaml -n memphis"
      sh "rm -rf memphis-k8s"
      sh "git clone --branch staging git@github.com:Memphis-OS/memphis-k8s.git"
      sh 'helm install my-emphis memphis-k8s/helm/memphis --set analytics="false" --create-namespace --namespace memphis'
    }

    stage('Build docker image') {
	    sh "docker buildx build --push -t ${dockerImagesRepo}/${imageName}:${versionTag} --platform linux/amd64,linux/arm/v7,linux/arm64 ."
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
