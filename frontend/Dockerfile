FROM node:19.2-alpine

WORKDIR /app

COPY package.json package.json
COPY yarn.lock yarn.lock

RUN yarn

COPY . .

RUN npx svelte-kit sync

CMD ["yarn", "dev", "--host"]
