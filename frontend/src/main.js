import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./assets/main.css";

const app = createApp(App);

app.config.errorHandler = function (err, vm, info) {
    console.error("GLOBAL ERROR HANDLER:", err, info);
};

app.use(router);
app.mount("#app");