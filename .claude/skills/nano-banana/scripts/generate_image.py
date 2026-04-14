#!/usr/bin/env python3
"""
generate_image.py — Nano Banana 2 image generation for Grupo BECM / Polimentes
Usage: python generate_image.py --prompt "..." --filename "output.png" --resolution 4K
       python generate_image.py --prompt "edit: make sky darker" --input-image "original.png" --filename "edited.png"
"""
import argparse
import os
import sys
import base64
import json
from pathlib import Path

def get_api_key():
    """Get API key from args, env, or APIs.env file"""
    for env_file in ['APIs.env', '.env', '../APIs.env']:
        env_path = Path(env_file)
        if env_path.exists():
            with open(env_path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith('OPENROUTER_API_KEY='):
                        return ('openrouter', line.split('=', 1)[1].strip('"\''))
                    if line.startswith('GEMINI_API_KEY='):
                        return ('gemini', line.split('=', 1)[1].strip('"\''))
    key = os.environ.get('OPENROUTER_API_KEY')
    if key:
        return ('openrouter', key)
    key = os.environ.get('GEMINI_API_KEY')
    if key:
        return ('gemini', key)
    return None, None

def generate_via_gemini(prompt, resolution, api_key, input_image_b64=None):
    import urllib.request
    url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image:generateContent?key={api_key}"
    parts = [{"text": f"Generate an image: {prompt}. High quality, {resolution} resolution."}]
    if input_image_b64:
        parts.insert(0, {"inline_data": {"mime_type": "image/png", "data": input_image_b64}})
    payload = json.dumps({
        "contents": [{"parts": parts}],
        "generationConfig": {"responseModalities": ["TEXT", "IMAGE"]}
    }).encode('utf-8')
    req = urllib.request.Request(url, data=payload, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=60) as response:
        result = json.loads(response.read())
        candidates = result.get('candidates', [])
        if not candidates:
            raise ValueError(f"No candidates in response: {json.dumps(result)[:300]}")
        parts = candidates[0].get('content', {}).get('parts', [])
        for part in parts:
            if 'inlineData' in part:
                img_data = part['inlineData']['data']
                return base64.b64decode(img_data)
        text_parts = [p.get('text', '') for p in parts if 'text' in p]
        raise ValueError(f"No image in response. Text: {' '.join(text_parts)[:300]}")

def generate_via_openrouter(prompt, resolution, api_key, input_image_b64=None):
    import urllib.request
    messages = [{"role": "user", "content": prompt}]
    if input_image_b64:
        messages[0]["content"] = [
            {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{input_image_b64}"}},
            {"type": "text", "text": prompt}
        ]
    payload = json.dumps({
        "model": "google/gemini-flash-1-5",
        "messages": messages,
        "max_tokens": 1024,
    }).encode('utf-8')
    req = urllib.request.Request(
        "https://openrouter.ai/api/v1/chat/completions",
        data=payload,
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "HTTP-Referer": "https://github.com/P0l1-0825-001-MX/poliagents-hub",
            "X-Title": "PoliAgents Hub - Grupo BECM"
        }
    )
    with urllib.request.urlopen(req) as response:
        result = json.loads(response.read())
        content = result['choices'][0]['message']['content']
        if isinstance(content, list):
            for item in content:
                if item.get('type') == 'image_url':
                    img_data = item['image_url']['url']
                    if img_data.startswith('data:image'):
                        return base64.b64decode(img_data.split(',')[1])
        raise ValueError(f"No image in response. Got: {str(content)[:200]}")

def main():
    parser = argparse.ArgumentParser(description='Generate images with Nano Banana 2')
    parser.add_argument('--prompt', required=True, help='Image description or edit instruction')
    parser.add_argument('--filename', required=True, help='Output filename (e.g., hero.png)')
    parser.add_argument('--resolution', default='1K', choices=['1K', '2K', '4K'])
    parser.add_argument('--input-image', help='Input image for editing (optional)')
    parser.add_argument('--api-key', help='Override API key')
    args = parser.parse_args()

    if args.api_key:
        provider, api_key = 'openrouter', args.api_key
    else:
        provider, api_key = get_api_key()

    if not api_key:
        print("No API key found.")
        print("   Run: echo 'GEMINI_API_KEY=...' >> APIs.env")
        print("   Or:  echo 'OPENROUTER_API_KEY=sk-or-v1-...' >> APIs.env")
        sys.exit(1)

    input_image_b64 = None
    if args.input_image:
        with open(args.input_image, 'rb') as f:
            input_image_b64 = base64.b64encode(f.read()).decode()

    output_path = Path(args.filename)

    print(f"Generando imagen con Nano Banana 2...")
    print(f"   Prompt: {args.prompt[:80]}...")
    print(f"   Resolution: {args.resolution}")
    print(f"   Provider: {provider}")

    try:
        if provider == 'gemini':
            img_bytes = generate_via_gemini(args.prompt, args.resolution, api_key, input_image_b64)
        else:
            img_bytes = generate_via_openrouter(args.prompt, args.resolution, api_key, input_image_b64)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, 'wb') as f:
            f.write(img_bytes)
        print(f"Imagen guardada: {output_path.absolute()}")
        print(f"   Size: {len(img_bytes) / 1024:.1f} KB")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
