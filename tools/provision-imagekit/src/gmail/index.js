import {
  authorize,
  getProfile,
  listMessages,
  getMessage,
  normalizeMessage,
  searchMessages,
} from "./gmail.js";

async function main() {
  console.log("Connecting to Gmail...");

  await authorize();

  console.log("Gmail API connected successfully.");

  const profile = await getProfile();

  console.log("\n=== GMAIL PROFILE ===");

  console.log({
    email: profile.emailAddress,
    messages: profile.messagesTotal,
    threads: profile.threadsTotal,
  });

  console.log("\n=== LIST MESSAGES ===");

  const result = await listMessages({
    query: "newer_than:1d",
    maxResults: 10,
  });

  console.log(`Found ${result.messages.length} messages`);

  for (const item of result.messages) {
    console.log({
      id: item.id,
      threadId: item.threadId,
    });
  }

  /*
   * Get details of the first email.
   */
  if (result.messages.length > 0) {
    const first = await getMessage(result.messages[0].id);

    const email = normalizeMessage(first);

    console.log("\n=== FIRST EMAIL ===");

    console.log({
      id: email.id,
      threadId: email.threadId,
      from: email.from,
      to: email.to,
      subject: email.subject,
      date: email.date,
      snippet: email.snippet,
    });

    console.log("\n=== TEXT CONTENT ===");

    console.log(email.text);

    console.log("\n=== HTML CONTENT ===");

    console.log(email.html);
  }

  /*
   * Search example.
   */
  console.log("\n=== SEARCH ===");

  const search = await searchMessages({
    query: "newer_than:1d",
    maxResults: 10,
  });

  for (const email of search.messages) {
    console.log({
      id: email.id,
      from: email.from,
      to: email.to,
      subject: email.subject,
      date: email.date,
    });
  }
}

main().catch((error) => {
  console.error("\nGmail API FAILED");

  console.error(error.message);

  if (error.response?.data) {
    console.error(error.response.data);
  }

  process.exit(1);
});
