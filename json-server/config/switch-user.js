/**
 * Middleware to allow the current user to be switched in order to test role-based conditional access.
 * Set the x-test-user-id cookie to the user ID to switch to.
 */
module.exports = (req, res, next) => {
    if (req.path === "/users/current") {
        let userID = req.headers?.cookie?.match(/x-test-user-id=(?<userID>[^;]+);?/)?.groups
            .userID;

        if (!userID) {
            userID = "1"; // default test user
        }

        req.url = `/users/${userID}`;
    }

    next();
};
