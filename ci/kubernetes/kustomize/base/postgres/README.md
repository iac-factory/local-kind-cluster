```bash
kubectl --namespace development exec -it services/postgres -- psql -h localhost -U api-service-user --password -p 5432 postgres
```
