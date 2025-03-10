FROM node:18-alpine AS build
WORKDIR /app
COPY package.json ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build

RUN ls -la /app

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
CMD ["nginx", "-g", "daemon off;"]
