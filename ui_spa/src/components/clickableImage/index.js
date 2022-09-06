import { Button } from 'antd';
import React from 'react';

const ClickableImage = (props) => {
    const { image, style, onClick } = props;

    return (
        <Button type="link" style={style ? style : null} onClick={onClick} {...props}>
            <img src={image} alt={image}></img>
        </Button>
    );
};

export default ClickableImage;
