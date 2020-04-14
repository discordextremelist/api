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

const express = require("express");
const rateLimit = require("rate-limiter-flexible");
const redis = require("redis");
const router = express.Router();
const redisClient = redis.createClient();

const settings = require("../../settings.json");

const apiRateSettings = new rateLimit.RateLimiterRedis({
    storeClient: redisClient,
    points: 75,
    duration: 1,
    blockDuration: 10,
    keyPrefix: "ratelimiter"
});

router.get('*', async (req, res, next) => {
    if (!settings.bypassedIPs.includes(req.clientIp)) {
        const token = req.headers["authorization"];

        if (!token) return res.status(401).json({ error: true, status: 401, message: "Unauthorised" });

        const bot = await req.app.db.collection("bots").findOne({ token: token });

        if (!bot || bot.token !== token) return res.status(403).json({ error: true, status: 403, message: "Invalid \"Authorization\" Header" });

        if (bot.status.verified === true) {
            apiRateSettings.consume(token, 5)
                .then(_ => { next(); })
                .catch(x => {
                    res.status(429).json({
                        error: true,
                        status: 429,
                        message: "Too Many Requests",
                        retry: Math.round(x.msBeforeNext / 1000) || 1
                });
            });
        } else {
            apiRateSettings.consume(token, 15)
                .then(_ => { next(); })
                .catch(x => {
                    res.status(429).json({
                        error: true,
                        status: 429,
                        message: "Too Many Requests",
                        retry: Math.round(x.msBeforeNext / 1000) || 1
                });
            });
        } 
    } else next();
});

router.get("/stats", async (req, res) => {
   const bots = await req.app.db.collection("bots").find().toArray();
   const users = await req.app.db.collection("users").find().toArray();
   const servers = await req.app.db.collection("servers").find().toArray();
   
   return res.status(200).json({ 
       error: false, 
       status: 200, 
       stats: { 
           servers: {
               total: servers.length
           },
           bots: { 
               total: bots.length, 
               approved: bots.filter(b => b.status.approved).length, 
               verified: bots.filter(b => b.status.verified).length 
               
           }, 
           users: {
               total: users.length,
               verified: users.filter(u => u.rank.verified).length,
               testers: users.filter(u => u.rank.tester).length,
               translators: users.filter(u => u.rank.translator).length,
               staff: {
                   total: users.filter(u => u.rank.mod).length,
                   mods: users.filter(u => u.rank.mod).length - users.filter(u => u.rank.assistant).length,
                   assistants: users.filter(u => u.rank.assistant).length - users.filter(u => u.rank.admin).length,
                   admins: users.filter(u => u.rank.admin).length
               }
           }
       } 
   });
});

router.get("/user/:id", async (req, res) => {
    const user = await req.app.db.collection("users").findOne({ "id": req.params.id }, { projection: { _id: 0, status: 0, preferences: 0, locale: 0, staffTracking: 0 } });
    
    if (!user) return res.status(404).json({
        error: true,
        message: "Unknown User",
        status: 404
    });

    res.status(200).json({ error: false, status: 200, user });
});

router.get("/server/:id", async (req, res) => {
    const server = await req.app.db.collection("servers").findOne({ "id": req.params.id }, { projection: { _id: 0, links: 0 } });
    
    if (!server) return res.status(404).json({
        error: true,
        message: "Unknown Server",
        status: 404
    });

    res.status(200).json({ error: false, status: 200, server });
});

router.get("/bot/:id", async (req, res) => {
    const bot = await req.app.db.collection("bots").findOne({ id: req.params.id }, { projection: { _id: 0, token: 0, modNotes: 0, votes: 0, "status.pendingVerification": 0 } });
    
    if (!bot) return res.status(404).json({
        error: true,
        message: "Unknown Bot",
        status: 404
    });

    res.status(200).json({ error: false, status: 200, bot });
});

router.post("/bot/:id", async (req, res) => {
    if (settings.api.bypass.ips.includes(req.clientIp)) {
        const token = req.headers["authorization"];

        if (!token) return res.status(401).json({ error: true, status: 401, message: "Unauthorised" });

        const bot = await req.app.db.collection("bots").findOne({ token: token });

        if (!bot || bot.token !== token) return res.status(403).json({ error: true, status: 403, message: "Invalid \"Authorization\" Header" });
    }
        
    const bot = await req.app.db.collection("bots").findOne({ "id": req.params.id });
    if (!bot) return res.status(404).json({ error: true, message: "Unknown Bot", status: 404 });

    if (!req.body.guildCount) return res.status(400).json({
        error: true,
        message: "guildCount (int) is Required",
        status: 400
    });

    if (isNaN(req.body.guildCount)) return res.status(400).json({
        error: true,
        message: "guildCount (int) is Required",
        status: 400
    });

    req.app.db.collection("users").updateOne({ id: req.user.id }, 
        { $set: {
            serverCount: Number(req.body.guildCount)
        }
    });

    res.status(200).json({ error: false, message: "Updated", status: 200 });
});

module.exports = router;
