FROM node:22-alpine as build

WORKDIR /app

COPY package.json ./
COPY package-lock.json ./
COPY babel.config.js ./

RUN npm install

COPY src ./src
COPY static ./static
COPY docs ./docs
COPY docusaurus.config.js ./
COPY sidebars.js ./

RUN npm run build

EXPOSE 3000

CMD ["npm", "run", "serve"]
