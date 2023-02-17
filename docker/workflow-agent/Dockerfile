FROM prefecthq/prefect:2.8-python3.10

RUN pip install prefect_aws
WORKDIR /app
COPY entrypoint.sh /app/entrypoint.sh

CMD ["sh","/app/entrypoint.sh"]
