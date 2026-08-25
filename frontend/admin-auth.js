const adminPoolData = {
    UserPoolId: "PASTE_ADMIN_USER_POOL_ID_HERE",
    ClientId: "PASTE_ADMIN_APP_CLIENT_ID_HERE"
};

const adminUserPool =
    new AmazonCognitoIdentity.CognitoUserPool(adminPoolData);

const adminForm =
    document.getElementById("adminLoginForm");

const adminError =
    document.getElementById("adminError");

const adminCreateAccount =
    document.getElementById("adminCreateAccount");

const adminForgotPassword =
    document.getElementById("adminForgotPassword");


// ===============================
// ADMIN SIGN UP
// ===============================

adminCreateAccount.addEventListener("click", function() {

    const phone = prompt(
        "Enter admin mobile number with country code, e.g. +919876543210"
    );

    if (!phone) return;

    const password = prompt(
        "Create admin password:"
    );

    if (!password) return;

    const phoneAttribute =
        new AmazonCognitoIdentity.CognitoUserAttribute({
            Name: "phone_number",
            Value: phone.trim()
        });

    adminUserPool.signUp(
        phone.trim(),
        password,
        [phoneAttribute],
        null,

        function(error) {

            if (error) {
                alert(error.message);
                return;
            }

            alert(
                "Verification code sent to admin mobile number."
            );

            verifyAdmin(phone.trim());
        }
    );
});


function verifyAdmin(phone) {

    const code = prompt(
        "Enter verification code:"
    );

    if (!code) return;

    const cognitoUser =
        new AmazonCognitoIdentity.CognitoUser({
            Username: phone,
            Pool: adminUserPool
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
                "Admin account verified successfully."
            );
        }
    );
}


// ===============================
// ADMIN LOGIN
// ===============================

adminForm.addEventListener("submit", function(event) {

    event.preventDefault();

    const phone =
        document.getElementById("adminPhone")
            .value.trim();

    const password =
        document.getElementById("adminPassword")
            .value;

    adminError.textContent = "Signing in...";

    const authenticationDetails =
        new AmazonCognitoIdentity.AuthenticationDetails({
            Username: phone,
            Password: password
        });

    const cognitoUser =
        new AmazonCognitoIdentity.CognitoUser({
            Username: phone,
            Pool: adminUserPool
        });

    cognitoUser.authenticateUser(
        authenticationDetails,
        {
            onSuccess: function(result) {

                sessionStorage.setItem(
                    "dfsAdminAuthenticated",
                    "true"
                );

                sessionStorage.setItem(
                    "dfsAdminIdToken",
                    result.getIdToken().getJwtToken()
                );

                sessionStorage.setItem(
                    "dfsAdminAccessToken",
                    result.getAccessToken().getJwtToken()
                );

                sessionStorage.setItem(
                    "dfsAdminPhone",
                    phone
                );

                adminError.textContent =
                    "Admin login successful ✓";

                setTimeout(() => {
                    window.location.href =
                        "admin-dashboard.html";
                }, 400);
            },

            onFailure: function(error) {

                console.error(error);

                sessionStorage.removeItem(
                    "dfsAdminAuthenticated"
                );

                adminError.textContent =
                    "Invalid admin mobile number or password.";
            }
        }
    );
});


// ===============================
// ADMIN FORGOT PASSWORD
// ===============================

adminForgotPassword.addEventListener(
    "click",
    function() {

        const phone = prompt(
            "Enter registered admin mobile number:"
        );

        if (!phone) return;

        const cognitoUser =
            new AmazonCognitoIdentity.CognitoUser({
                Username: phone.trim(),
                Pool: adminUserPool
            });

        cognitoUser.forgotPassword({

            onFailure: function(error) {
                alert(
                    error.message ||
                    "Unable to reset admin password."
                );
            },

            inputVerificationCode: function() {

                const code = prompt(
                    "Enter verification code:"
                );

                if (!code) return;

                const newPassword = prompt(
                    "Enter new admin password:"
                );

                if (!newPassword) return;

                cognitoUser.confirmPassword(
                    code,
                    newPassword,
                    {
                        onSuccess: function() {
                            alert(
                                "Admin password reset successful."
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
