FROM ubuntu:latest

ARG FURIOSA_IAM_KEY
ARG FURIOSA_IAM_PWD

RUN apt update && \
    apt install -y ca-certificates apt-transport-https gnupg wget gpg pip

RUN mkdir -p /etc/apt/keyrings && \
    wget -q -O- https://archive.furiosa.ai/furiosa-apt-key.gpg | gpg --dearmor | tee /etc/apt/keyrings/furiosa-apt-key.gpg > /dev/null

RUN tee -a /etc/apt/auth.conf.d/furiosa.conf > /dev/null <<EOT
  machine archive.furiosa.ai
  login ${FURIOSA_IAM_KEY}
  password ${FURIOSA_IAM_PWD}
EOT
RUN chmod 400 /etc/apt/auth.conf.d/furiosa.conf

RUN tee -a /etc/apt/sources.list.d/furiosa.list <<EOT
deb [arch=amd64 signed-by=/etc/apt/keyrings/furiosa-apt-key.gpg] https://archive.furiosa.ai/ubuntu jammy restricted
EOT
RUN apt update && apt install -y furiosa-libhal-warboy furiosa-compiler libonnxruntime && apt clean && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --upgrade pip setuptools wheel requests && \
    pip install --no-cache-dir 'furiosa-models'

COPY e2e/inference_pod/license_free_cat.jpg /license_free_cat.jpg
COPY e2e/inference_pod/run.py /run.py

CMD ["python3", "/run.py"]
