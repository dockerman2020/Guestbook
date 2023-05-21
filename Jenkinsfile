pipeline{
    agent {
        kubernetes {
            cloud 'kubernetes'
            // defaultContainer 'maven'
            idleMinutes 5
            namespace 'jenkins-worker'
            yaml '''
                apiVersion: v1
                kind: Pod
                spec:
                    containers:
                    - name: maven
                      image: dockerman2002/e2e-dev:0.41.0@sha256:10e291ab783b24060b6f9f60f15051c6672db855b82e240acd04728063eb315c
                      command:
                      - cat
                      tty: true
                      volumeMounts:
                      - name: docker-socket
                        mountPath: /var/run/docker.sock
                    volumes:
                    - name: docker-socket
                      hostPath:
                        path: /var/run/docker.sock
                        type: Socket
            '''
        }
	 }
	 
    tools {
        jdk 'java17'
        maven 'maven3'
    }
    
    environment {
        APP_NAME = "e2e-pipeline"
        RELEASE = "1.0.0"
        DOCKER_USER = "dockerman2002"
        DOCKER_PASS = 'dockerhub'
        IMAGE_NAME = "${DOCKER_USER}" + "/" + "${APP_NAME}"
        IMAGE_TAG = "${RELEASE}-${BUILD_NUMBER}"
    //     // JENKINS_API_TOKEN = credentials("JENKINS_API_TOKEN")

    }
    stages{
        stage("Cleanup Workspace"){
            steps {
                cleanWs()
            }
        }
    
        stage("Checkout from SCM"){
            steps {
                git branch: 'main', credentialsId: 'github-1', url: 'https://github.com/dockerman2020/Guestbook.git'
            }
        }

    stage('SonarQube Analysis') {
        def scannerHome = tool 'SonarScanner';
        withSonarQubeEnv() {
        sh "${scannerHome}/bin/sonar-scanner"
        }
    }

        stage("Build Application"){
            steps {
                script {
                sh 'mvn clean package'
                }
            }
        }

        stage("Test Application"){
            steps {
                sh  'mvn test'
            }
        }
        
        stage("Sonarqube Analysis") {
            steps {
                script {
                    withSonarQubeEnv(credentialsId: 'guestbook-sonar') {
                        sh 'mvn sonar:sonar'
                    }
                }
            }
        }

        stage("Quality Gate") {
            steps {
                script {
                    waitForQualityGate abortPipeline: false, credentialsId: 'guestbook-sonar'
                }
            }
        }

        stage("Build & Push Docker Image") {
            steps {
                container('maven') {
                    script {
                        docker.withRegistry('',DOCKER_PASS) {
                            docker_image = docker.build "${IMAGE_NAME}"
                        }
                        docker.withRegistry('',DOCKER_PASS) {
                            docker_image.push("${IMAGE_TAG}")
                            docker_image.push('latest')
                        }
                    }
                }
            }
        }

        stage("Trivy Scan") {
            steps {
                container('maven') {
                    sh 'trivy image ${IMAGE_NAME}:${IMAGE_TAG} --scanners vuln --severity HIGH,CRITICAL,MEDIUM -f json'
                }
            }
        }
        
        stage ('Cleanup Artifacts') {
            steps {
                script {
                    container('maven') {
                    sh ('docker run -v /var/run/docker.sock:/var/run/docker.sock docker rmi ${IMAGE_NAME}:${IMAGE_TAG} && docker rmi ${IMAGE_NAME}:latest')
                    }
                }
            }
        }

        // stage("Trigger CD Pipeline") {
        //     steps {
        //         script {
        //             sh "curl -v -k --user admin:${JENKINS_API_TOKEN} -X POST -H 'cache-control: no-cache' -H 'content-type: application/x-www-form-urlencoded' --data 'IMAGE_TAG=${IMAGE_TAG}' 'https://jenkins.dev.dman.cloud/job/gitops-complete-pipeline/buildWithParameters?token=gitops-token'"
        //         }
        //     }
        // }
    }

    // post {
    //     failure {
    //         emailext body: '''${SCRIPT, template="groovy-html.template"}''', 
    //                 subject: "${env.JOB_NAME} - Build # ${env.BUILD_NUMBER} - Failed", 
    //                 mimeType: 'text/html',to: "dmistry@yourhostdirect.com"
    //         }
    //      success {
    //            emailext body: '''${SCRIPT, template="groovy-html.template"}''', 
    //                 subject: "${env.JOB_NAME} - Build # ${env.BUILD_NUMBER} - Successful", 
    //                 mimeType: 'text/html',to: "dmistry@yourhostdirect.com"
    //       }      
    // }
    post {
        always {
            script {
                def status = currentBuild.result ?: 'UNKNOWN'
                def color
                switch (status) {
                    case 'SUCCESS':
                        color = 'good'
                        break
                    case 'FAILURE':
                        color = 'danger'
                        break
                    default:
                        color = 'warning'
                }
                slackSend(channel: 'C057R7SPGLT', message: "Update Deployment ${status.toLowerCase()} for ${env.JOB_NAME} ${env.BUILD_NUMBER} - ${env.BUILD_URL}",
                iconEmoji: ':jenkins:', color: color)
            }
        }
    }
}
