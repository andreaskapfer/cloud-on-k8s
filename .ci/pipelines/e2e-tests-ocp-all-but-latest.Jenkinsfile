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
        timeout(time: 50, unit: 'HOURS')
    }

    environment {
        VAULT_ADDR = credentials('vault-addr')
        VAULT_ROLE_ID = credentials('vault-role-id')
        VAULT_SECRET_ID = credentials('vault-secret-id')
        GCLOUD_PROJECT = credentials('k8s-operators-gcloud-project')
    }

    stages {
        stage('Load common scripts') {
            steps {
                script {
                    lib = load ".ci/common/tests.groovy"
                }
            }
        }
        // latest 4.x is taken care of by a separate job
        // individual build jobs will report error but we want to run all jobs in all cases so we catch the errors
        // and fail only the stage
        stage("4.7.x "){
            steps {
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                    build job: 'cloud-on-k8s-e2e-tests-ocp',
                                    parameters: [
                                        string(name: 'JKS_PARAM_OPERATOR_IMAGE', value: JKS_PARAM_OPERATOR_IMAGE),
                                        string(name: 'OCP_VERSION', value: "4.7.59"),
                                        string(name: 'branch_specifier', value: GIT_COMMIT)
                                    ],
                                    wait: true
                }
            }
        }
        stage("4.8.x "){
            steps {
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                    build job: 'cloud-on-k8s-e2e-tests-ocp',
                                    parameters: [
                                        string(name: 'JKS_PARAM_OPERATOR_IMAGE', value: JKS_PARAM_OPERATOR_IMAGE),
                                        string(name: 'OCP_VERSION', value: "4.8.50"),
                                        string(name: 'branch_specifier', value: GIT_COMMIT)
                                    ],
                                    wait: true
                }
            }
        }
        stage("4.9.x "){
            steps {
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                    build job: 'cloud-on-k8s-e2e-tests-ocp',
                                    parameters: [
                                        string(name: 'JKS_PARAM_OPERATOR_IMAGE', value: JKS_PARAM_OPERATOR_IMAGE),
                                        string(name: 'OCP_VERSION', value: "4.9.48"),
                                        string(name: 'branch_specifier', value: GIT_COMMIT)
                                    ],
                                    wait: true
                }
            }
        }
        stage("4.10.x "){
            steps {
                catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
                    build job: 'cloud-on-k8s-e2e-tests-ocp',
                                    parameters: [
                                        string(name: 'JKS_PARAM_OPERATOR_IMAGE', value: JKS_PARAM_OPERATOR_IMAGE),
                                        string(name: 'OCP_VERSION', value: "4.10.34"),
                                        string(name: 'branch_specifier', value: GIT_COMMIT)
                                    ],
                                    wait: true
                }
            }
        }
    }
}

