# Переменная для отслеживания успешности операций
has_failed=false

POD_NAME=$(kubectl get pods -n $JENKINS_NAMESPACE -l $JENKINS_LABEL -o jsonpath='{.items[0].metadata.name}')

# Переменная для хранения имени бэкап файла
BACKUP_FILE_NAME="jenkins-backup-$(date +'%Y%m%d%H%M%S').tar.gz"

# Создание архива директории
echo -e "Start archiving the jenkins home directory at $(date +'%d-%m-%Y %H:%M:%S')."
kubectl exec -n $JENKINS_NAMESPACE $POD_NAME -- tar czf /tmp/$BACKUP_FILE_NAME $JENKINS_HOME
echo -e "Starting to copy the archive the jenkins home directory at $(date +'%d-%m-%Y %H:%M:%S')."
kubectl cp -n $JENKINS_NAMESPACE $POD_NAME:/tmp/$BACKUP_FILE_NAME /tmp/$BACKUP_FILE_NAME

# Перемещение архива в S3
if awsoutput=$(aws --endpoint-url $AWS_S3_ENDPOINT_URL s3 cp /tmp/$BACKUP_FILE_NAME s3://$AWS_BUCKET_NAME$AWS_BUCKET_BACKUP_PATH/$BACKUP_FILE_NAME 2>&1)
then
    echo -e "Jenkins backup successfully uploaded to s3 at $(date +'%d-%m-%Y %H:%M:%S')."
    # Проверка и удаление старых бэкапов, если их количество превышает установленное значение
    if [ $(aws --endpoint-url $AWS_S3_ENDPOINT_URL s3 ls s3://$AWS_BUCKET_NAME$AWS_BUCKET_BACKUP_PATH/ | wc -l) -gt $BACKUP_DEPTH ]
    then
        # Нахождение и удаление самого старого бэкапа
        oldest_backup=$(aws --endpoint-url $AWS_S3_ENDPOINT_URL s3 ls s3://$AWS_BUCKET_NAME$AWS_BUCKET_BACKUP_PATH/ | sort | head -n 1 | awk '{print $4}')
        aws --endpoint-url $AWS_S3_ENDPOINT_URL s3 rm s3://$AWS_BUCKET_NAME$AWS_BUCKET_BACKUP_PATH/$oldest_backup
        echo -e "Removed oldest backup: $oldest_backup from s3."
    fi
else
    echo -e "Jenkins backup failed to upload at $(date +'%d-%m-%Y %H:%M:%S'). Error: $awsoutput" | tee -a /tmp/kubernetes-s3-directory-backup.log
    has_failed=true
fi

# Удаление временного файла бэкапа
rm -f /tmp/$BACKUP_FILE_NAME