import './style.scss';
import React, { useState } from 'react';
import playVideoIcon from '../../assets/images/playVideoIcon.svg';
import ReactPlayer from 'react-player';

const VideoPlayer = (props) => {
    const { url } = props;
    const [playState, setPlayState] = useState(false);

    return (
        <div className="video-container">
            <div className="video">
                <ReactPlayer
                    controls={true}
                    playing={playState}
                    light={true}
                    playIcon={
                        <div onClick={() => setPlayState(true)}>
                            <img alt="play-video-icon" src={playVideoIcon} />
                        </div>
                    }
                    height="220px"
                    width="320px"
                    url={url}
                ></ReactPlayer>
            </div>
        </div>
    );
};

export default VideoPlayer;
