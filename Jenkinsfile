pipeline {
    agent { label 'pr' }

    options {
        timestamps ()
        timeout(time: 1, unit: 'HOURS')
    }

    environment {
        GOMEMLIMIT = "5GiB"
    }

    stages {
        stage('Validate commit') {
            steps {
                script {
                    def CHANGE_REPO = sh (script: "basename -s .git `git config --get remote.origin.url`", returnStdout: true).trim()
                    build job: '/Utils/Validate-Git-Commit', parameters: [
                        string(name: 'Repo', value: "${CHANGE_REPO}"),
                        string(name: 'Branch', value: "${env.CHANGE_BRANCH}"),
                        string(name: 'Commit', value: "${GIT_COMMIT}")
                    ]
                }
            }
        }

        stage('Check Go sources formatting') {
            steps {
                sh 'diff=`gofmt -s -d .`; echo "$diff"; test -z "$diff"'
            }
        }
        stage('Lint') {
            steps {
                //TODO remove binary
                sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.0.2'
                sh 'golangci-lint run ./...'
            }
        }
        stage('Run go tests') {
            steps {
                sh 'go mod tidy'
                sh 'go test ./... -timeout 30m'
            }
        }
    }
}
