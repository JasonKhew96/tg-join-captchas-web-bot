window.Telegram.WebApp.ready();

// setup theme
let checkThemeFunc = () => {
  let body = document.querySelector("body");
  if (window.Telegram.WebApp.colorScheme == "dark") {
    body.classList.remove("bootstrap");
    body.classList.add("bootstrap-dark");
  } else {
    body.classList.remove("bootstrap-dark");
    body.classList.add("bootstrap");
  }
};
checkThemeFunc();

Telegram.WebApp.onEvent("themeChanged", () => {
  checkThemeFunc();
});

let submitFunc = (ids) => {
  let answers = [];
  for (let id of ids) {
    answers.push({
      id: id,
      answer: document.getElementById("question" + id).value,
    });
  }
  let body = {
    init_data: Telegram.WebApp.initData,
    answers: answers,
    version: Telegram.WebApp.version,
    platform: Telegram.WebApp.platform,
  };
  fetch("/api/submit", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  })
    .then((response) => response.json())
    .then((json) => {
      window.Telegram.WebApp.close();
    });
};

let validateData = {
  init_data: Telegram.WebApp.initData,
};

fetch("/api/validate", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify(validateData),
})
  .then((response) => response.json())
  .then((json) => {
    if (json.status) {
      document.getElementsByTagName("title").textContent = json.title;
      document.getElementById("title").textContent = json.title;

      let ids = [];
      for (let i = 0; i < json.questions.length; i++) {
        let inputGroup = document.createElement("div");
        inputGroup.classList.add("input-group");
        inputGroup.classList.add("mb-3");

        let span = document.createElement("span");
        span.classList.add("input-group-text");
        span.id = "answer" + json.questions[i].id;
        span.textContent = "Answer";

        inputGroup.appendChild(span);

        let input = document.createElement("input");
        input.type = "text";
        input.classList.add("form-control");
        input.id = "question" + json.questions[i].id;

        ids.push(json.questions[i].id);

        inputGroup.appendChild(input);

        let question = document.createElement("div");

        let label = document.createElement("label");
        label.classList.add("form-label");
        label.textContent = json.questions[i].question;

        question.appendChild(label);
        question.appendChild(inputGroup);

        document.querySelector("#questions").appendChild(question);
      }

      let submit = document.createElement("button");
      submit.type = "button";
      submit.classList.add("btn");
      submit.classList.add("btn-primary");
      submit.textContent = "Submit";

      submit.onclick = (e) => {
        submit.disabled = true;
        submitFunc(ids);
      };

      document.querySelector("#submit").appendChild(submit);
    } else {
      let alert = document.createElement("div");
      alert.classList.add("alert");
      alert.classList.add("alert-danger");
      alert.setAttribute("role", "alert");
      alert.textContent = json.message;

      document.querySelector("#alert").appendChild(alert);
    }
  });
