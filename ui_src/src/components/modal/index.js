// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect } from 'react';
import { Modal } from 'antd';

import Button from '../button';

const TransitionsModal = ({
    height,
    width,
    rBtnText,
    lBtnText,
    rBtnDisabled,
    lBtnDisabled,
    header,
    isLoading,
    open = false,
    displayButtons = true,
    lBtnClick,
    clickOutside,
    rBtnClick,
    children,
    hr = false,
    className,
    zIndex = null,
    keyListener = true,
    onPressEnter = () => {}
}) => {
    const contentStyle = {
        height: height,
        overflowY: 'auto',
        borderTop: hr ? '1px solid #EAEAEA' : null
    };

    useEffect(() => {
        const keyDownHandler = (event) => {
            if (event.key === 'Enter') {
                if (displayButtons) {
                    event.preventDefault();
                    rBtnClick();
                } else onPressEnter();
            }
        };
        if (open && keyListener) {
            document.addEventListener('keydown', keyDownHandler);
        }
        return () => {
            document.removeEventListener('keydown', keyDownHandler);
        };
    }, [open]);

    return (
        <Modal
            wrapClassName={className || 'modal-wrapper'}
            title={header}
            open={open}
            width={width || 'fit-content'}
            onCancel={() => clickOutside()}
            bodyStyle={contentStyle}
            centered
            destroyOnClose={true}
            zIndex={zIndex}
            footer={
                displayButtons
                    ? [
                          <div key="left" className="btnContainer">
                              <button className="cancel-button" disabled={lBtnDisabled} onClick={() => lBtnClick()}>
                                  {lBtnText}
                              </button>
                              <Button
                                  className="modal-btn"
                                  width="83px"
                                  height="32px"
                                  placeholder={rBtnText}
                                  disabled={rBtnDisabled}
                                  colorType="white"
                                  radiusType="circle"
                                  backgroundColorType={'purple'}
                                  fontSize="12px"
                                  fontWeight="600"
                                  isLoading={isLoading}
                                  onClick={() => {
                                      rBtnClick();
                                  }}
                              />
                          </div>
                      ]
                    : null
            }
        >
            {children}
        </Modal>
    );
};

export default TransitionsModal;
