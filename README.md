# streamlining-backend

run go build
docker build -t <name>
docker tag <name> <core.harbor.domain/library/name>
docker push <core.harbor.domain/library/name>
kubectl apply -f .