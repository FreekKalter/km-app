from fabric.api import *
import os
import time
from path import path
from operator import attrgetter

env.ssh_config_path = '/var/lib/jenkins/.ssh/config'
env.use_ssh_config = True
#env.key_filename = '/var/lib/jenkins/.ssh/id_rsa'
env.hosts.extend(['fkalter@km-app.kalteronline.org'])

def localTest():
    killContainers(local)
    local('docker run  -v /home/fkalter/km/postgresdata:/data:rw\
                       -v /home/fkalter/km/log:/log:rw\
                       --name postgres\
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
    if not os.environ.has_key('NO_DOCKER_BUILD'):
        buildContainers()
        pushContainers()
    run("docker pull freekkalter/km-app")
    run("docker pull freekkalter/nginx")
    runProduction(run)

def manualDeploy():
    buildContainers('MANUL')
    pushContainers()
    run("docker pull freekkalter/km-app")
    run("docker pull freekkalter/nginx")
    runProduction(run, 'MANUAL')

def buildContainers(name=None):
    if not name:
        name = os.environ['BUILD_NUMBER']
    local("make app/km minify")
    local('mkdir -p nginx/static/js')
    local('mkdir -p nginx/static/css')
    local('mkdir -p nginx/static/img')
    local('cp -R app/img nginx/static')
    local('cp -R app/js nginx/static')
    local('cp -R app/css nginx/static')
    local('cp -R app/partials nginx/static')

    local("cp config-production.yml app/config.yml")

    latest =  getLatestBuildNr('local')
    print latest
    with open('Dockerfile.template', 'r') as i, open('Dockerfile', 'w') as o:
        for l in i.xreadlines():
            o.write(l.replace('BASE', str(latest)))
    local("docker build -t freekkalter/km-app:{} .".format(name))
    local("docker build -t freekkalter/nginx:deploy nginx")

def pushContainers():
    local("docker push freekkalter/km-app")
    local("docker push freekkalter/nginx")

def killContainers(method):
    with settings(hide('warnings'), warn_only=True):
        local('pkill km$')
        method("docker kill km_production")
        method("docker rm km_production")
        method("docker kill nginx")
        method("docker rm nginx")
        method("docker kill postgres")
        method("docker rm postgres")

def runProduction(method, buildName=None):
    if not buildName:
        buildName = os.environ['BUILD_NUMBER']
    killContainers(method)
    method("docker run --name km_production \
                           -v /home/fkalter/km/postgresdata:/data:rw\
                           -v /home/fkalter/km/log:/log\
                           -d -p 4001:4001 \
                           freekkalter/km-app:{} /usr/bin/supervisord".format(buildName) )

    method("docker run -d -p 443:443 --link km_production:app --name nginx\
                          -v /home/fkalter/km/ssl:/etc/nginx/conf:ro \
                          -v /home/fkalter/km/log:/log\
                          freekkalter/nginx:deploy /start_nginx")

def rollback():
    # find latest buildnumber on remote, default = the build before the last one
    buildNumber = int(prompt('Rever to buildnumber: ', validate=int, default=getLatestBuildNr('run')-1))
    if buildNumber < 0:
        buildNumber = bn+buildNumber
    runProduction(run, buildNumber)

def getSqlDump(directory):
    run('docker run -v /home/fkalter/backup:/backup:rw --link km_production:main\
                    freekkalter/km-app:{} /backup.sh'.format(getLatestBuildNr('run')))
    get('/home/fkalter/backup/backup.sql', directory)

def pullProductionData():
    local('mkdir -p backup')
    getSqlDump('./backup')

    # import into local running container
    buildnr = getLatestBuildNr('local')
    runProduction(local, buildnr)
    time.sleep(2)
    local('docker run -v /home/fkalter/github/km/backup:/backup:rw --link km_production:main freekkalter/km-app:{} /restore.sh'.format(buildnr))

# call backup from cronjob/jenkins
def backup():
    getSqlDump('.')
    # tar file and move to folder with format backup-{date}.sql
    local('mkdir -p ~/km-backup')
    local("tar -czf ~/km-backup/backup_`date +%d-%m-%Y.tar.gz` ./backup.sql")
    local("rm ./backup.sql")

def restore(backuparchive):
    local("tar -C /tmp --overwrite -xzf "+backuparchive)

    uploaded = put('/tmp/backup.sql', '/home/fkalter/backup/backup.sql')
    if uploaded.failed:
        print uploaded.failed + 'failed to upload'
        return
    run('docker run -v /home/fkalter/backup:/backup:rw --link km_production:main freekkalter/km-app:{} /restore.sh'.format(getLatestBuildNr('run')))

def getLatestBuildNr(method):
    try:
        if method == 'local':
            return int(local("docker images | awk '{ if(match($2, /^[0-9]+$/)) print $2}' | sort -n | tail -n1", capture=True))
        else:
            return int(run("docker images | awk '{ if(match($2, /^[0-9]+$/)) print $2}' | sort -n | tail -n1"))
    except Exception as e:
        return 'base'

