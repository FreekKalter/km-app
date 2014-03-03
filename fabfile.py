from fabric.api import *
import os
import time

env.use_ssh_config = True
env.ssh_config_path = '/var/lib/jenkins/.ssh/config'
env.hosts.extend(['fkalter@paris'])

def localTest():
    killContainers(local)
    local('docker run  -v /home/fkalter/km/postgresdata:/data:rw\
                       -v /home/fkalter/km/log:/log:rw\
                       -name postgres\
                       -d -p 5432:5432\
                       freekkalter/postgres-supervisord:km /usr/bin/supervisord')
    local('make app/km')
    local('cp ./config-testing.yml app/config.yml')
    with lcd('./app'):
        local('./km &')

def localDeploy():
    buildName = 'local'
    buildContainers(buildName)
    runProduction(local, buildName)

def deploy():
    buildNr = os.environ['BUILD_NUMBER']
    buildContainers(buildNr)
    #pushContainers()
    run("docker pull freekkalter/km")
    run("docker pull freekkalter/nginx")
    runProduction(run, buildNr)

def buildContainers(buildNr):
    local("make app/km minify")
    local('mkdir -p nginx/static/js')
    local('mkdir -p nginx/static/css')
    local('cp app/js/master.js nginx/static/js/')
    local('cp app/css/main.min.css nginx/static/css/')
    local('cp app/favicon.ico nginx/static')
    local('cp -R app/partials nginx/static')

    local("cp config-production.yml app/config.yml")
    local("docker build -t freekkalter/km:{} .".format(buildNr))
    local("docker build -t freekkalter/nginx:deploy nginx")

def pushContainers():
    local("docker push freekkalter/km")
    local("docker push freekkalter/nginx")

def killContainers(method):
    with settings(hide('warnings'), warn_only=True):
        local('pkill km')
        method("docker kill km_production")
        method("docker rm km_production")
        method("docker kill nginx")
        method("docker rm nginx")
        method("docker kill postgres")
        method("docker rm postgres")

def runProduction(method, buildName):
    killContainers(method)
    method("docker run -name km_production \
                           -v /home/fkalter/km/postgresdata:/data:rw\
                           -v /home/fkalter/km/log:/log\
                           -d -p 4001:4001 \
                           freekkalter/km:{} /usr/bin/supervisord".format(buildName) )

    method("docker run -d -p 443:443 -link km_production:app -name nginx\
                                  -v /home/fkalter/km/ssl:/etc/nginx/conf:ro \
                                  freekkalter/nginx:deploy /start_nginx")

def rollback():
    # find latest buildnumber on remote, default = the build before the last one
    buildNumber = int(prompt('Rever to buildnumber: ', validate=int, default=getLatestBuildNr()-1))
    if buildNumber < 0:
        buildNumber = bn+buildNumber
    runProduction(run, buildNumber)

def getSqlDump(directory):
    run('docker run -v /home/fkalter/backup:/backup:rw -link km_production:main freekkalter/km:{} /backup.sh'.format(getLatestBuildNr()))
    get('/home/fkalter/backup/backup.sql', directory)

def pullProductionData():
    local('mkdir -p backup')
    getSqlDump('./backup')

    # import into local running container
    runProduction(local, getLatestBuildNr())
    time.sleep(2)
    local('docker run -v /home/fkalter/github/km/backup:/backup:rw -link km_production:main freekkalter/km:deploy /restore.sh')

# call backup from cronjob/jenkins
def backup():
    getSqlDump('.')
    # tar file and move to folder with format backup-{date}.sql
    local('mkdir -p ~/km-backup')
    local("tar -czf ~/km-backup/backup_`date +%d-%m-%Y.tar.gz` ./backup.sql")
    local("rm ./backup.sql")

def getLatestBuildNr():
    return int(run("docker images | awk '{ if(match($2, /^[0-9]+$/)) print $2}' | sort | tail -n1"))
