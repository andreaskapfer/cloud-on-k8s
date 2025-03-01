// This library overrides the default checkout behavior to enable sleep+retries if there are errors
// Added to help overcome some recurring github connection issues
@Library('apm@current') _

def failedTests = []
def lib

pipeline {

    agent {
        label 'linux'
    }

    options {
        timeout(time: 600, unit: 'MINUTES')
        skipDefaultCheckout(true)
    }

    environment {
        VAULT_ADDR = credentials('vault-addr')
        VAULT_ROLE_ID = credentials('vault-role-id')
        VAULT_SECRET_ID = credentials('vault-secret-id')
        GCLOUD_PROJECT = credentials('k8s-operators-gcloud-project')
    }

    stages {
        stage('Checkout, stash source code and load common scripts') {
            steps {
                checkout scm
                stash allowEmpty: true, name: 'source', useDefaultExcludes: false
                script {
                    lib = load ".ci/common/tests.groovy"
                }
            }
        }
        stage('Run tests for different k8s versions in GKE') {
            parallel {
                stage("1.21") {
                    agent {
                        label 'linux'
                    }
                    steps {
                        unstash "source"
                        script {
                            runWith(lib, failedTests, '1.21', "eck-gke21-${BUILD_NUMBER}-e2e")
                        }
                    }
                }
                stage("1.22") {
                    agent {
                        label 'linux'
                    }
                    steps {
                        unstash "source"
                        script {
                            runWith(lib, failedTests, '1.22', "eck-gke22-${BUILD_NUMBER}-e2e")
                        }
                    }
                }
                stage("1.23") {
                    agent {
                        label 'linux'
                    }
                    steps {
                        unstash "source"
                        script {
                            runWith(lib, failedTests, '1.23', "eck-gke23-${BUILD_NUMBER}-e2e")
                        }
                    }
                }
                stage("1.24") {
                    agent {
                        label 'linux'
                    }
                    steps {
                        unstash "source"
                        script {
                            runWith(lib, failedTests, '1.24', "eck-gke24-${BUILD_NUMBER}-e2e")
                        }
                    }
                }
            }
        }
    }

    post {
        unsuccessful {
            script {
                if (params.SEND_NOTIFICATIONS) {
                    Set<String> filter = new HashSet<>()
                    filter.addAll(failedTests)
                    def msg = lib.generateSlackMessage("E2E tests for different k8s versions in GKE failed!", env.BUILD_URL, filter)

                    slackSend(
                        channel: '#eck',
                        color: 'danger',
                        message: msg,
                        tokenCredentialId: 'cloud-ci-slack-integration-token',
                        failOnError: true
                    )
                }
            }
        }
        cleanup {
            script {
                clusters = [
                    "eck-gke21-${BUILD_NUMBER}-e2e",
                    "eck-gke22-${BUILD_NUMBER}-e2e",
                    "eck-gke23-${BUILD_NUMBER}-e2e",
                    "eck-gke24-${BUILD_NUMBER}-e2e"
                ]
                for (int i = 0; i < clusters.size(); i++) {
                    build job: 'cloud-on-k8s-e2e-cleanup',
                        parameters: [string(name: 'JKS_PARAM_GKE_CLUSTER', value: clusters[i])],
                        wait: false
                }
            }
            cleanWs()
        }
    }
}

def runWith(lib, failedTests, clusterVersion, clusterName) {
    sh ".ci/setenvconfig e2e/gke-k8s-versions $clusterVersion $clusterName"
    script {
        env.SHELL_EXIT_CODE = sh(returnStatus: true, script: 'make -C .ci get-test-artifacts TARGET=ci-e2e ci')

        sh 'make -C .ci TARGET=e2e-generate-xml ci'
        junit "e2e-tests.xml"

        if (env.SHELL_EXIT_CODE != 0) {
            failedTests.addAll(lib.getListOfFailedTests())
            googleStorageUpload bucket: "gs://devops-ci-artifacts/jobs/$JOB_NAME/$BUILD_NUMBER",
                credentialsId: "devops-ci-gcs-plugin",
                pattern: "*.zip",
                sharedPublicly: true,
                showInline: true
        }

        sh 'exit $SHELL_EXIT_CODE'
    }
}
