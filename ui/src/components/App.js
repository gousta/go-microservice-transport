import React from "react";

export default class App extends React.Component {
  render() {
    console.log("window.location.href", window.location.href);
    console.log("process.env", process.env);
    return <div>Welcome to another react app</div>;
  }
}
