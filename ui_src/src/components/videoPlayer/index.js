import './style.scss';
import React, { useState } from 'react';
import playVideoIcon from '../../assets/images/playVideoIcon.svg';
import ReactPlayer from 'react-player';

const VideoPlayer = ({ url, err }) => {
    const [playState, setPlayState] = useState(false);

    return (
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
            onError={() => err(true)}
        ></ReactPlayer>
    );
};

export default VideoPlayer;
