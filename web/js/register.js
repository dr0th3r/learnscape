registrationForm = document.getElementById("registration-form")
registerSchool = document.getElementById("register-school")
registerAdmin = document.getElementById("register-admin")
gotoRegisterAdmin = document.getElementById("goto-register-admin")
goBack = document.getElementById("go-back")
errorDiv = document.getElementById("result")

gotoRegisterAdmin.addEventListener("click", () => {
	registerSchool.style.display = "none"
	registerAdmin.style.display = "flex"
})

goBack.addEventListener("click", () => {
	registerSchool.style.display = "flex"
	registerAdmin.style.display = "none"
})

registrationForm.addEventListener("htmx:responseError", (e) => {
	errorDiv.innerHTML = "<div class='text-red-600 text-center pb-2' id='result'>Error: " + e.detail.xhr.response + "</div>";
})
