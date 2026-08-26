const adminPoolData = {
    UserPoolId: "ap-south-1_RdvZbNKYZ",
    ClientId: "562c4caevoa8c45g3b9iqvgqf4"
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

adminForm.addEventListener("submit", async function(event) {
    event.preventDefault();

    const phone =
        document.getElementById("adminPhone").value.trim();

    const password =
        document.getElementById("adminPassword").value;

    adminError.textContent = "Signing in...";

    try {
        const response = await fetch(
            "https://cognito-idp.ap-south-1.amazonaws.com/",
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/x-amz-json-1.1",
                    "X-Amz-Target":
                        "AWSCognitoIdentityProviderService.InitiateAuth"
                },
                body: JSON.stringify({
                    AuthFlow: "USER_PASSWORD_AUTH",
                    ClientId: "562c4caevoa8c45g3b9iqvgqf4",
                    AuthParameters: {
                        USERNAME: phone,
                        PASSWORD: password
                    }
                })
            }
        );

        const data = await response.json();

        if (!response.ok) {
            console.error("Admin login failed:", data);

            adminError.textContent =
                data.message ||
                data.__type ||
                "Invalid admin mobile number or password.";

            return;
        }

        if (
            data.ChallengeName === "NEW_PASSWORD_REQUIRED"
        ) {
            adminError.textContent =
                "Admin must set a new password first.";
            return;
        }

        if (!data.AuthenticationResult) {
            adminError.textContent =
                "Authentication could not be completed.";
            return;
        }

        sessionStorage.setItem(
            "dfsAdminAuthenticated",
            "true"
        );

        sessionStorage.setItem(
            "dfsAdminIdToken",
            data.AuthenticationResult.IdToken
        );

        sessionStorage.setItem(
            "dfsAdminAccessToken",
            data.AuthenticationResult.AccessToken
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

    } catch (error) {
        console.error("Admin network error:", error);

        adminError.textContent =
            "Unable to connect to Cognito. Please try again.";
    }
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
