// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useState } from 'react';
import ReactPlayer from 'react-player';

import playVideoIcon from '../../assets/images/playVideoIcon.svg';
import Img404 from '../../assets/images/404.svg';

const VideoPlayer = ({ url, bgImg }) => {
    const [playState, setPlayState] = useState(false);
    const [isOffline, setIsOffline] = useState(false);

    return isOffline ? (
        <img className="not-connected" src={Img404} alt="not connected" />
    ) : (
        <ReactPlayer
            className="video-player"
            controls={true}
            playing={playState}
            light={true}
            playIcon={
                <div onClick={() => setPlayState(true)}>
                    <img alt="play-video-icon" src={playVideoIcon} />
                </div>
            }
            height="250px"
            width="445px"
            url={url}
            onError={() => setIsOffline(true)}
            style={{ backgroundImage: `url(${bgImg})`, backgroundRepeat: 'no-repeat', backgroundSize: 'cover' }}
        ></ReactPlayer>
    );
};

export default VideoPlayer;
