FROM python:3.10

WORKDIR /app

RUN pip install mlflow==2.1.1
RUN pip install psycopg2-binary==2.9.5
COPY entrypoint.sh /app/entrypoint.sh

EXPOSE 5000

CMD ["sh","/app/entrypoint.sh"]