/**
 * Middleware to add/remove headers from response
 */
module.exports = (req, res, next) => {
    res.removeHeader('X-Powered-By');
    next();
};
