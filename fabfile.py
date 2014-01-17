from fabric.api import *
import os

env.use_ssh_config = True
env.ssh_config_path = '/var/lib/jenkins/.ssh/config'
env.hosts.extend(['fkalter@km-app.dyndns.org'])

def localTest():
    killLocalContainers(False)
    local('docker run  -v /home/fkalter/km/postgresdata:/data:rw\
                       -v /home/fkalter/km/log:/log:rw\
                       -name postgres\
                       -d -p 5432:5432\
                       freekkalter/postgres-supervisord:km /usr/bin/supervisord')
    local('make app/km')
    with settings(warn_only=True):
        local('pkill km')
    local('cp ./config-testing.yml app/config.yml')
    with lcd('./app'):
        local('./km &')

def localDeploy():
    buildName = 'local'
    buildContainers(buildName)
    runProduction(False, buildName)

def deploy():
    buildNr = os.environ['BUILD_NUMBER']
    buildContainers(buildNr)
    pushContainers()
    run("docker pull freekkalter/km")
    runProduction(True, buildNr)

def buildContainers(buildNr):
    local("make app/km minify")
    local("cp config-production.yml app/config.yml")
    local("docker build -t freekkalter/km:{} .".format(buildNr))
    local("docker build -t freekkalter/nginx:deploy nginx")

def pushContainers():
    local("docker push freekkalter/km")
    local("docker push freekkalter/nginx")

def killLocalContainers(remote):
    if remote:
        command = run
    else:
        command = local
    with settings(warn_only=True):
        command("docker kill km_production")
        command("docker rm km_production")
        command("docker kill nginx")
        command("docker rm nginx")
        command("docker kill postgres")
        command("docker rm postgres")

def runProduction(remote, buildName):
    if remote:
        command = run
    else:
        command = local
    killLocalContainers(remote)
    command("docker run -name km_production \
                           -v /home/fkalter/km/postgresdata:/data:rw\
                           -v /home/fkalter/km/log:/log\
                           -d -p 4001:4001 \
                           freekkalter/km:{} /usr/bin/supervisord".format(buildName) )

    command("docker run -d -p 443:443 -link km_production:app -name nginx\
                                  -v /home/fkalter/km/ssl:/etc/nginx/conf:ro \
                                  freekkalter/nginx:deploy /start_nginx")

def rollback():
    # find latest buildnumber on remote, default = the build before the last one
    buildNumber = int(prompt('Rever to buildnumber: ', validate=int, default=getLatestBuildNr()-1))
    if buildNumber < 0:
        buildNumber = bn+buildNumber
    runProduction(True, buildNumber)

def getSqlDump(directory):
    run('docker run -v /home/fkalter/backup:/backup:rw -link km_production:main freekkalter/km:{} /backup.sh'.format(getLatestBuildNr()))
    get('/home/fkalter/backup/backup.sql', directory)

def pullProductionData():
    local('mkdir -p backup')
    getSqlDump('./backup')

    # import into local running container
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
