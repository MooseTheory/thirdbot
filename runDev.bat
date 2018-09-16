@ECHO OFF

docker-compose --file docker-compose-dev.yml build
docker-compose --file docker-compose-dev.yml up