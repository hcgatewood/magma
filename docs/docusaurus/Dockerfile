FROM node:8.11.4

WORKDIR /app/website

COPY package.json /app/website/package.json
RUN yarn install

CMD yarn install && yarn start
