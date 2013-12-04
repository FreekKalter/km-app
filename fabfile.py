from fabric.api import local, run, cd

def deploy():
    local("docker build -t freekkalter/km:deploy .")
    local("docker push freekkalter/km")
    #TODO: find clean way to kill old running container (cidfile?)
    run("docker pull freekkalter/km")
    run("docker run -d -p 4001:4001 -v /home/fkalter/km/postgresdata:/data:rw\
                                    -v /home/fkalter/km/log:/log\
                                    freekkalter/km:deploy /usr/bin/supervisord")
