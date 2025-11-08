<?php
/**
 * AIWF PHP Client Usage Example
 *
 * This example demonstrates how to use the auto-generated PHP HTTP client
 * to interact with AIWF AI agents.
 *
 * Prerequisites:
 * 1. Start AIWF server: aiwf serve -f php-translator-example.yaml
 * 2. Set OPENAI_API_KEY environment variable
 * 3. Run this script: php php-translator-usage.php
 */

require_once __DIR__ . '/generated/client.php';

use AIWFClient\AIWFClient;
use AIWFClient\TranslateRequest;

// Create client (connects to local server by default)
$client = new AIWFClient(
    baseURL: 'http://127.0.0.1:8080',
    apiKey: null // Set API key if server requires authentication
);

try {
    // Create translation request
    $request = new TranslateRequest(
        target_lang: 'en',
        text: 'Привет, мир! Это пример использования AIWF.'
    );

    // Call translator agent
    echo "🔄 Translating...\n";
    $response = $client->translator($request);

    // Display results
    echo "\n✅ Translation complete!\n\n";
    echo "Original: {$request->text}\n";
    echo "Translated: {$response->translated}\n";
    echo "Source Language: {$response->source_lang}\n";
    echo "Confidence: " . ($response->confidence * 100) . "%\n";

    echo "\n" . str_repeat("=", 50) . "\n";

    // Another example - translate to different language
    $request2 = new TranslateRequest(
        target_lang: 'es',
        text: 'Hello, world!'
    );

    echo "\n🔄 Translating to Spanish...\n";
    $response2 = $client->translator($request2);

    echo "\n✅ Translation complete!\n\n";
    echo "Original: {$request2->text}\n";
    echo "Translated: {$response2->translated}\n";
    echo "Confidence: " . ($response2->confidence * 100) . "%\n";

} catch (\Exception $e) {
    echo "❌ Error: " . $e->getMessage() . "\n";
    echo "\nMake sure AIWF server is running:\n";
    echo "  aiwf serve -f examples/php-translator-example.yaml\n";
    exit(1);
}
