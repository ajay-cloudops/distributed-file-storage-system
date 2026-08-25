const poolData = {
    UserPoolId: "ap-south-1_PvnFOJAx5",
    ClientId: "38l9spdp5n3a3u181i6cmelpq5"
};

const userPool =
    new AmazonCognitoIdentity.CognitoUserPool(poolData);

const loginForm =
    document.getElementById("loginForm");

const createAccountButton =
    document.getElementById("createAccountButton");

const forgotPasswordButton =
    document.getElementById("forgotPasswordButton");

const errorBox =
    document.getElementById("loginError");


// ===============================
// REAL COGNITO LOGIN
// ===============================

loginForm.addEventListener("submit", function(event) {
    event.preventDefault();

    const email =
        document.getElementById("email").value.trim();

    const password =
        document.getElementById("password").value;

    errorBox.textContent = "Signing in...";

    const authenticationDetails =
        new AmazonCognitoIdentity.AuthenticationDetails({
            Username: email,
            Password: password
        });

    const userData = {
        Username: email,
        Pool: userPool
    };

    const cognitoUser =
        new AmazonCognitoIdentity.CognitoUser(userData);

    cognitoUser.authenticateUser(
        authenticationDetails,
        {
            onSuccess: function(result) {

                sessionStorage.setItem(
                    "dfsAuthenticated",
                    "true"
                );

                sessionStorage.setItem(
                    "dfsIdToken",
                    result.getIdToken().getJwtToken()
                );

                sessionStorage.setItem(
                    "dfsAccessToken",
                    result.getAccessToken().getJwtToken()
                );

                sessionStorage.setItem(
                    "dfsUserEmail",
                    email
                );

                errorBox.textContent =
                    "Login successful ✓";

                document.body.classList.add("leaving");

                setTimeout(() => {
                    window.location.href =
                        "storage-select.html";
                }, 400);
            },

            onFailure: function(error) {

                console.error("Login failed:", error);

                sessionStorage.removeItem(
                    "dfsAuthenticated"
                );

                sessionStorage.removeItem(
                    "dfsIdToken"
                );

                sessionStorage.removeItem(
                    "dfsAccessToken"
                );

                errorBox.textContent =
                    "Invalid email or password.";
            }
        }
    );
});


// ===============================
// CREATE ACCOUNT
// ===============================

createAccountButton.addEventListener(
    "click",
    function() {

        const email =
            prompt("Enter your email address:");

        if (!email) {
            return;
        }

        const password =
            prompt(
                "Create your password:"
            );

        if (!password) {
            return;
        }

        const emailAttribute =
            new AmazonCognitoIdentity.CognitoUserAttribute({
                Name: "email",
                Value: email
            });

        userPool.signUp(
            email,
            password,
            [emailAttribute],
            null,

            function(error) {

                if (error) {
                    alert(error.message);
                    return;
                }

                alert(
                    "Verification code sent to your email."
                );

                verifyAccount(email);
            }
        );
    }
);


function verifyAccount(email) {

    const code =
        prompt(
            "Enter verification code:"
        );

    if (!code) {
        return;
    }

    const cognitoUser =
        new AmazonCognitoIdentity.CognitoUser({
            Username: email,
            Pool: userPool
        });

    cognitoUser.confirmRegistration(
        code,
        true,

        function(error) {

            if (error) {
                alert(error.message);
                return;
            }

            alert(
                "Email verified. You can now sign in."
            );
        }
    );
}


// ===============================
// FORGOT PASSWORD
// ===============================

forgotPasswordButton.addEventListener(
    "click",
    function() {

        const email = prompt(
            "Enter your registered email address:"
        );

        if (!email) {
            return;
        }

        const cognitoUser =
            new AmazonCognitoIdentity.CognitoUser({
                Username: email.trim(),
                Pool: userPool
            });

        cognitoUser.forgotPassword({

            onSuccess: function() {
                alert(
                    "Password changed successfully. You can now sign in."
                );
            },

            onFailure: function(error) {
                alert(
                    error.message ||
                    "Unable to reset password."
                );
            },

            inputVerificationCode: function() {

                alert(
                    "A verification code has been sent to your email."
                );

                const code = prompt(
                    "Enter the verification code:"
                );

                if (!code) {
                    return;
                }

                const newPassword = prompt(
                    "Enter your new password:"
                );

                if (!newPassword) {
                    return;
                }

                cognitoUser.confirmPassword(
                    code,
                    newPassword,
                    {
                        onSuccess: function() {
                            alert(
                                "Password reset successful ✓"
                            );
                        },

                        onFailure: function(error) {
                            alert(
                                error.message ||
                                "Password reset failed."
                            );
                        }
                    }
                );
            }
        });
    }
);
