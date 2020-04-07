/*
Discord Extreme List - Discord's unbiased list.

Copyright (C) 2020 Cairo Mitchell-Acason

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

const createError = require("http-errors");
const requestIp = require("request-ip");
const express = require("express");
const logger = require("morgan");
const app = express();

const settings = require("./settings.json");

console.log("Mongo: Connection opening...");
const { MongoClient } = require("mongodb");
let db;
new Promise((resolve, reject) => {
    MongoClient.connect(settings.mongo, { useUnifiedTopology: true }, (error, mongo) => {
        if (error) return reject(error);
        db = mongo.db("del");
        console.log("Mongo: Connection established! Released deadlock as a part of startup...");
        resolve();
    });
}).then(async () => {
    app.db = db;

    app.set("trust proxy", true);

    app.use(logger("combined"));

    app.use(express.json());
    app.use(express.urlencoded({ extended: false }));

    app.use(requestIp.mw());

    app.use("/", require("./src/Routes/index.js"));
    app.use("/v1", require("./src/Routes/v1-REVISED.js"));

    app.use((req, res, next) => {
        next(createError(404));
    });

    app.use((err, req, res, next) => {
        res.locals.message = err.message;
        res.locals.error = req.app.get("env") === "development" ? err : {};

        res.status(err.status || 500);

        if (err.message === "Not Found") {
            return res.status(404).json({ error: true, status: 404, message: "Unknown Endpoint" });
        }
    });
});

module.exports = app;