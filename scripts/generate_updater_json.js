const fs = require('fs');
const path = require('path');

const eventPath = process.env.GITHUB_EVENT_PATH;
const eventData = JSON.parse(fs.readFileSync(eventPath, 'utf8'));
const version = eventData.release.tag_name.replace(/^v/, '');
const releaseNotes = eventData.release.body;
const pubDate = eventData.release.published_at;

const platforms = {};

function addPlatform(name, signaturePath, url) {
    try {
        const signature = fs.readFileSync(signaturePath, 'utf8');
        platforms[name] = {
            signature: signature.trim(),
            url: url
        };
        console.log(`Added ${name}`);
    } catch (e) {
        console.error(`Failed to read signature for ${name}: ${e.message}`);
    }
}

// Walk through artifacts
// Expected structure: ./artifacts/{platform}/{filename}
const artifactsDir = process.env.ARTIFACTS_DIR || '.';
const baseUrl = `https://github.com/orhaniscoding/goconnect/releases/latest/download`;

fs.readdirSync(artifactsDir).forEach(file => {
    if (file.endsWith('.sig')) {
        const signaturePath = path.join(artifactsDir, file);
        const assetName = file.replace('.sig', '');
        const url = `${baseUrl}/${assetName}`;

        // Infer platform from filename logic used in release.yml
        // Linux: .AppImage.tar.gz
        // Windows: .zip
        // MacOS: .app.tar.gz

        if (file.includes('linux') || file.includes('AppImage')) {
            addPlatform('linux-x86_64', signaturePath, url);
        } else if (file.includes('windows') || file.includes('setup.zip')) {
            addPlatform('windows-x86_64', signaturePath, url);
        } else if (file.includes('darwin')) {
            if (file.includes('aarch64')) {
                addPlatform('darwin-aarch64', signaturePath, url);
            } else {
                addPlatform('darwin-x86_64', signaturePath, url);
            }
        }
    }
});

const manifest = {
    version,
    notes: releaseNotes,
    pub_date: pubDate,
    platforms
};

fs.writeFileSync('latest.json', JSON.stringify(manifest, null, 2));
console.log('Generated latest.json');
