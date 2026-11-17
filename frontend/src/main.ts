import { mount } from "svelte";
import "@fontsource-variable/inter";
import "./style.css";
import App from "./App.svelte";
mount(App, { target: document.getElementById("app")! });
