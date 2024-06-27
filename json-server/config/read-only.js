/**
 * Read-only middleware to prevent the db.json being modified
 */
module.exports = (req, res, next) => {
    if (["POST", "PUT", "PATCH"].includes(req.method)) {
        req.method === "POST" ? res.status(201) : res.status(200);
        req.method = "GET";
    }
    next();
};
