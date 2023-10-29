FROM node:lts-alpine

# make the 'app' folder the current working directory
WORKDIR /app
COPY package*.json ./
RUN yarn

COPY . .

CMD [ "yarn", "dev", "--host"]