from fabric.api import *

env.ssh_config_path = '/home/fkalter/.ssh/config'
env.use_ssh_config = True
env.hosts.extend( ['fkalter@km-app.dyndns.org'])

def deploy():
    prepare()
    remote()

def prepare():
    local("make production")
    local("docker build -t freekkalter/km:deploy .")
    local("docker push freekkalter/km")

def remote():
    cidfile = '/home/fkalter/.km.cidfile'
    run("docker pull freekkalter/km")
    run("docker kill `cat {}`".format(cidfile))
    run("rm {}".format(cidfile))
    run("docker run -cidfile={}\
                     -v /home/fkalter/km/postgresdata:/data:rw\
                     -v /home/fkalter/km/log:/log\
                     -d -p 4001:4001\
                     freekkalter/km:deploy /usr/bin/supervisord".format(cidfile))
