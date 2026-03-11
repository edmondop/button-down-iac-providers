import pulumi
import pulumi_buttondown as buttondown

# Read your account info
account = buttondown.get_account()
pulumi.export("account_username", account.username)

# Create tags
engineering = buttondown.Tag("engineering",
    name="Engineering",
    color="#0066cc",
)

# Create an email draft
weekly_update = buttondown.Email("weekly-update",
    subject="Weekly Engineering Update",
    body="# This Week in Engineering\n\nHere's what happened this week...",
    description="Weekly engineering team update",
)

# Set up a webhook
webhook = buttondown.Webhook("new-subscriber",
    url="https://hooks.example.com/buttondown",
    event_types=["subscriber.created", "subscriber.updated"],
    description="Notify our CRM when subscribers change",
)
