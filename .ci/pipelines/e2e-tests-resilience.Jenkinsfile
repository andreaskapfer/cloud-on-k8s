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
        stage('Checkout and load common scripts') {
            steps {
                checkout scm
                script {
                    lib = load ".ci/common/tests.groovy"
                }
            }
        }
        stage('Run ECK resilience tests in GKE') {
            steps {
                script {
                    runWith(lib, failedTests, "eck-resilience-${BUILD_NUMBER}-e2e")
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
                    def msg = lib.generateSlackMessage("E2E resilience tests failed!", env.BUILD_URL, filter)

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
                build job: 'cloud-on-k8s-e2e-cleanup',
                    parameters: [string(name: 'JKS_PARAM_GKE_CLUSTER', value: "eck-jks-e2e-resilience-${BUILD_NUMBER}")],
                    wait: false
            }
            cleanWs()
        }
    }
}

def runWith(lib, failedTests, clusterName) {
    sh ".ci/setenvconfig e2e/resilience $clusterName"
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
