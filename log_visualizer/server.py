"""Server handles the logs visualization page"""

import io
import os
import glob

from waitress import serve
from flask import Flask, redirect, render_template, jsonify, Response
from flask_cors import CORS
import pandas as pd
import matplotlib.pyplot as plt

app = Flask(__name__)
app.config["DEBUG"] = True

# Allow access from the rulat-bot.duckdns.org domain
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
    """Health check endpoint"""
    return jsonify(status="healthy"), 200

@app.route('/graph/top-time-spent-players', methods=['GET'])
def top_time_spent_players():
    """
    Returns an image with a bar chart of the top 10 players by total session time.
    Uses 'connected' and 'disconnected' records from CSV files.
    """
    csv_files = glob.glob(os.path.join(CSV_DIR, '*.csv'))
    
    df_list = []
    for file in csv_files:
        df_part = pd.read_csv(file)
        df_list.append(df_part)
    
    if df_list:
        df = pd.concat(df_list, ignore_index=True)
    else:
        return render_template("dashboard.html")
    
    # Convert TimeStamp to datetime and sort by time
    df['TimeStamp'] = pd.to_datetime(df['TimeStamp'])
    df = df.sort_values('TimeStamp')
    
    # Dictionaries to store connection times and total session durations
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
    ax.set_ylabel('Session duration')
    ax.set_title('Top-Time-Spent Players')
    plt.xticks(rotation=45, ha='right')

    # Save the chart to a buffer as a JPEG image and return it as an image
    buf = io.BytesIO()
    plt.savefig(buf, format='jpeg', bbox_inches='tight')
    buf.seek(0)
    plt.close(fig)

    return Response(buf.getvalue(), mimetype='image/jpeg')

@app.route('/graph/top-counties-connected', methods=['GET'])
def top_counties_connected():
    """
    Returns an image with a pie chart showing the distribution of sessions by country.
    Only sessions with a duration longer than 30 seconds are considered.
    The analysis uses the 'Country' field from CSV files.
    """
    csv_files = glob.glob(os.path.join(CSV_DIR, '*.csv'))
    
    df_list = []
    for file in csv_files:
        df_part = pd.read_csv(file)
        df_list.append(df_part)
    
    if df_list:
        df = pd.concat(df_list, ignore_index=True)
    else:
        return render_template("dashboard.html")
    
    # Convert TimeStamp to datetime and sort by time
    df['TimeStamp'] = pd.to_datetime(df['TimeStamp'])
    df = df.sort_values('TimeStamp')

    # Dictionaries to store connection times and count sessions per country
    last_connected = {}
    country_sessions = {}

    # Process each record
    for _, row in df.iterrows():
        nick = row['NickName']
        action = row['Action']
        timestamp = row['TimeStamp']
        # If the Country field is missing or empty, set it to "Unknown"
        country = row['Country'] if pd.notna(row['Country']) and row['Country'] != "" else "Unknown"

        if action == 'connected':
            # Save connection time and country
            last_connected[nick] = (timestamp, country)
        elif action == 'disconnected':
            if nick in last_connected:
                start_time, start_country = last_connected[nick]
                session_time = (timestamp - start_time).total_seconds()
                # Count the session only if its duration is more than 30 seconds
                if session_time > 30:
                    country_sessions[start_country] = country_sessions.get(start_country, 0) + 1
                # Remove the connection record after disconnect
                del last_connected[nick]

    # Build a pie chart; if no data available, show "No data"
    if not country_sessions:
        fig, ax = plt.subplots(figsize=(6, 6))
        ax.text(0.5, 0.5, 'No data', horizontalalignment='center', verticalalignment='center')
    else:
        labels = list(country_sessions.keys())
        sizes = list(country_sessions.values())

        fig, ax = plt.subplots(figsize=(8, 8))
        ax.pie(sizes, labels=labels, autopct='%1.1f%%', startangle=90)
        ax.axis('equal')
        ax.set_title('Top Countries by Sessions (>30 sec)')

    # Save the pie chart to a buffer as a JPEG image and return it as an image
    buf = io.BytesIO()
    plt.savefig(buf, format='jpeg', bbox_inches='tight')
    buf.seek(0)
    plt.close(fig)

    return Response(buf.getvalue(), mimetype='image/jpeg')

# Start the server on port 5000
serve(app, host="0.0.0.0", port=5000)
