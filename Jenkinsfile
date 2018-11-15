#!/usr/bin/env groovy

pipeline {
	agent {
		docker {
			image 'golang:1.11'
			args '-u 0'
		 }
	}
	environment {
		GOBIN = '/usr/local/bin'
		DEBIAN_FRONTEND = 'noninteractive'
	}
	stages {
		stage('Bootstrap') {
			steps {
				echo 'Bootstrapping..'
				sh 'curl -sSL https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
				sh 'go get -v github.com/golang/lint/golint'
				sh 'go get -v github.com/tebeka/go2xunit'
				sh 'apt-get update && apt-get install -y build-essential'
			}
		}
		stage('Lint') {
			steps {
				echo 'Linting..'
				sh 'make lint | tee golint.txt || true'
				sh 'make vet | tee govet.txt || true'
			}
		}
		stage('Build') {
			steps {
				echo 'Building..'
				sh 'make'
				sh './bin/kapid version'
			}
		}
		stage('Test') {
			steps {
				echo 'Testing..'
				sh 'make test-xml-short'
			}
		}
		stage('Dist') {
			steps {
				echo 'Dist..'
				sh 'test -z "$(git diff --shortstat 2>/dev/null |tail -n1)" && echo "Clean check passed."'
				sh 'make check'
				sh 'make dist'
			}
		}
	}
	post {
		always {
			archiveArtifacts 'dist/*.tar.gz'
			junit allowEmptyResults: true, testResults: 'test/*.xml'
			warnings parserConfigurations: [[parserName: 'Go Lint', pattern: 'golint.txt'], [parserName: 'Go Vet', pattern: 'govet.txt']], unstableTotalAll: '0'
			cleanWs()
		}
	}
}
