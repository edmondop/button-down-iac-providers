import * as buttondown from "@pulumi/buttondown";

// Read your account info
const account = buttondown.getAccountOutput();
export const accountUsername = account.username;

// Create tags for organizing content
const engineering = new buttondown.Tag("engineering", {
    name: "Engineering",
    color: "#0066cc",
});

const announcements = new buttondown.Tag("announcements", {
    name: "Announcements",
    color: "#ff6600",
});

// Create a reusable snippet
const footerCta = new buttondown.Snippet("footer-cta", {
    identifier: "footer-cta",
    name: "Footer CTA",
    content: "Thanks for reading! Reply to this email if you have questions.",
    mode: "fancy",
});

// Create an email draft
const weeklyUpdate = new buttondown.Email("weekly-update", {
    subject: "Weekly Engineering Update",
    body: `# This Week in Engineering

Here's what happened this week...

{{ footer-cta }}`,
    description: "Weekly engineering team update",
});

// Set up a webhook for new subscribers
const webhook = new buttondown.Webhook("new-subscriber", {
    url: "https://hooks.example.com/buttondown",
    eventTypes: ["subscriber.created", "subscriber.updated"],
    description: "Notify our CRM when subscribers change",
});
