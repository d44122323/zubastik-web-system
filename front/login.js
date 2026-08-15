const form = document.getElementById("loginForm");
console.log("LOGIN JS LOADED");
form.addEventListener(
    "submit",
    async function(e) {
        e.preventDefault();
        console.log("FORM SUBMIT");
        const login =
            document.getElementById("login").value;
        const password =
            document.getElementById("password").value;
        const response = await fetch(
            "/admin/login",
            {
                method: "POST",
                credentials: "include",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    login: login,
                    password: password
                })
            }
        );
        console.log(
            "STATUS:",
            response.status
        );
if(response.ok){
    await new Promise(resolve => setTimeout(resolve, 200));
    window.location.replace("/dashboard.html");
}
        else {
            document.getElementById(
                "loginError"
            ).textContent =
                "Неверный логин или пароль";
        }
    }
);