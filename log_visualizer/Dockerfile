FROM python:3.12

WORKDIR /app
ADD ./requirements.txt /app/requirements.txt
RUN pip install -r requirements.txt
RUN pip install pylint
RUN python -m nltk.downloader stopwords wordnet

COPY . /app

EXPOSE $PORT

CMD ["python3", "server.py"]