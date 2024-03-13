/**
 * Custom Body Mapper middleware allows the body to be modified to add, remove,
 * or change values, as dome routes may otherwise update the db.json with
 * invalid data
 */
module.exports = (req, res, next) => {
    next();
};
