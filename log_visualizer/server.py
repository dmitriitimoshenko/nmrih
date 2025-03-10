"""Server handles the logs visualizaton page"""

import io
import base64
import os
import glob

from waitress import serve
from flask import Flask, redirect, render_template, jsonify, Response
from flask_cors import CORS
import pandas as pd
import matplotlib.pyplot as plt

app = Flask(__name__)
app.config["DEBUG"] = True

CORS(app, resources={r"/*": {"origins": "https://rulat-bot.duckdns.org"}})

CSV_DIR = '../data'

@app.after_request
def add_cors_headers(response):
    response.headers["Access-Control-Allow-Origin"] = "https://rulat-bot.duckdns.org"
    response.headers["Access-Control-Allow-Methods"] = "GET, POST, OPTIONS"
    response.headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
    return response

@app.route('/health-check', methods=['GET'])
def healthcheck():
    """health check"""
    return jsonify(status="healthy"), 200

@app.route('/')
def index():
    """will return the dashboard page"""
    return redirect('/dashboard')

@app.route('/dashboard', methods=['GET'])
def dashboard():
    """dashboard page has informative graphs on it"""

    csv_files = glob.glob(os.path.join(CSV_DIR, '*.csv'))
    
    df_list = []
    for file in csv_files:
        df_part = pd.read_csv(file)
        df_list.append(df_part)
    
    if df_list:
        df = pd.concat(df_list, ignore_index=True)
    else:
        return render_template("dashboard.html")
    
    df['TimeStamp'] = pd.to_datetime(df['TimeStamp'])
    df = df.sort_values('TimeStamp')
    
    # Dictionaries for holding connection times and total session durations
    last_connected = {}
    total_duration = {}

    # Iterate over all records
    for _, row in df.iterrows():
        nick = row['NickName']
        action = row['Action']
        timestamp = row['TimeStamp']

        if action == 'connected':
            # Record connection time
            last_connected[nick] = timestamp
        elif action == 'disconnected':
            # If the user was previously connected, calculate session duration
            if nick in last_connected:
                session_time = (timestamp - last_connected[nick]).total_seconds()
                total_duration[nick] = total_duration.get(nick, 0) + session_time
                # Remove the connection record after disconnect
                del last_connected[nick]

    # Sort players by total session time in descending order and select top 10
    sorted_players = sorted(total_duration.items(), key=lambda x: x[1], reverse=True)
    top_players = sorted_players[:10]
    players = [player for player, _ in top_players]
    durations = [duration for _, duration in top_players]

    # Build a bar chart
    fig, ax = plt.subplots(figsize=(10, 6))
    ax.bar(players, durations)
    ax.set_xlabel('NickName')
    ax.set_ylabel('Время сессии (сек.)')
    ax.set_title('Топ игроков по проведённому времени')
    plt.xticks(rotation=45, ha='right')

    # Save the plot to a buffer and encode it in base64 for HTML embedding
    buf = io.BytesIO()
    plt.savefig(buf, format='png', bbox_inches='tight')
    buf.seek(0)
    plt.close(fig)

    return Response(buf.getvalue(), mimetype='image/png')

serve(app, host="0.0.0.0", port=5000)
