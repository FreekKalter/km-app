from fabric.api import *
import os

env.use_ssh_config = True
env.ssh_config_path = '/var/lib/jenkins/.ssh/config'
env.hosts.extend( ['fkalter@km-app.dyndns.org'])

def deploy():
    prepareDeploy()
    remoteDeploy()

def prepareDeploy():
    buildNumber = os.environ['BUILD_NUMBER']
    local("make prepare-production")
    local("docker build -t freekkalter/km:{} .".format(buildNumber))
    local("docker push freekkalter/km")

def remoteDeploy():
    buildNumber = os.environ['BUILD_NUMBER']
    cidfile = '/home/fkalter/.km.cidfile'
    run("docker pull freekkalter/km")
    run("docker kill `cat {}`".format(cidfile))
    run("rm {}".format(cidfile))
    run("docker run -cidfile={}\
                     -v /home/fkalter/km/postgresdata:/data:rw\
                     -v /home/fkalter/km/log:/log\
                     -d -p 4001:4001\
                     freekkalter/km:{} /usr/bin/supervisord".format(cidfile, buildNumber))

# Rollback excepts a excact buildnumber or a negative offset to rollback that number builds
def rollback():
    # find latest buildnumber on remote, default = the build before the last one
    bn = int(run("docker images | awk '{ if(match($2, /^[0-9]+$/)) print $2}' | sort | tail -n1"))
    buildNumber = int(prompt('Rever to buildnumber: ', validate=int, default=bn-1))
    if buildNumber < 0:
        buildNumber = bn+buildNumber

    cidfile = '/home/fkalter/.km.cidfile'
    run("docker kill `cat {}`".format(cidfile))
    run("rm {}".format(cidfile))
    run("docker run -cidfile={}\
                     -v /home/fkalter/km/postgresdata:/data:rw\
                     -v /home/fkalter/km/log:/log\
                     -d -p 4001:4001\
                     freekkalter/km:{} /usr/bin/supervisord".format(cidfile, buildNumber))

