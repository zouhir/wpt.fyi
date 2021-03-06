FROM gcr.io/google-appengine/python

# Default WORKDIR: /home/vmagent/app (/app symlinks to this)

# Install curl & gcloud SDK.
RUN apt-get update -q
RUN apt-get install -qy curl
RUN curl -s https://sdk.cloud.google.com > install-gcloud.sh
RUN bash install-gcloud.sh --disable-prompts --install-dir=/opt > /dev/null
ENV PATH=/opt/google-cloud-sdk/bin:$PATH
RUN gcloud config set disable_usage_reporting false

# Setup and activate virtualenv.
RUN virtualenv --no-download /env -p python3
ENV VIRTUAL_ENV /env
ENV PATH /env/bin:$PATH

# Install Python dependencies.
ADD requirements.txt /app/
RUN pip install -r requirements.txt

ADD . /app/
# The number of workers should always be 2: one for processing tasks, the other
# for responding health checks. Scale the service by increasing the number of
# instances instead.
CMD exec gunicorn --bind :$PORT --timeout 3600 --workers 2 main:app
