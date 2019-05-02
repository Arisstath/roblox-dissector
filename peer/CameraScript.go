package peer

// CameraScript is the default Lua script to be replicated to
// server clients
const CameraScript = `
-- Sala Roblox Network Suite
-- Default Server Camera Script

local InfoHint = Instance.new("Hint", workspace);
local PerFrameMoveDelta = 1/60;
local MoveAmp = 30;
local PerFrameRotationDelta = 1/60;
local RotationAmp = 90;

local Camera = workspace.CurrentCamera;
local UIS = game:GetService "UserInputService";
local RunService = game:GetService "RunService";

-- Reset orientation
Camera.CFrame = CFrame.new(Camera.CFrame.Position);

RunService:BindToRenderStep("salaCamera", 200, function(delta) 
	InfoHint.Text = ("Sala Camera: Move Amplifier %d, Rotation Amplifier %d°"):format(MoveAmp, RotationAmp)
	if UIS:IsKeyDown(Enum.KeyCode.W) then
		Camera.CFrame = Camera.CFrame + Camera.CFrame.LookVector * PerFrameMoveDelta * MoveAmp;
	elseif UIS:IsKeyDown(Enum.KeyCode.S) then
		Camera.CFrame = Camera.CFrame + Camera.CFrame.LookVector * (-PerFrameMoveDelta) * MoveAmp;
	end
	if UIS:IsKeyDown(Enum.KeyCode.R) then
		Camera.CFrame = Camera.CFrame + Camera.CFrame.UpVector * PerFrameMoveDelta * MoveAmp;
	elseif UIS:IsKeyDown(Enum.KeyCode.F) then
		Camera.CFrame = Camera.CFrame + Camera.CFrame.UpVector * (-PerFrameMoveDelta) * MoveAmp;
	end
	if UIS:IsKeyDown(Enum.KeyCode.A) then
		Camera.CFrame = Camera.CFrame * CFrame.fromOrientation(0, math.rad(PerFrameRotationDelta * RotationAmp), 0);
	elseif UIS:IsKeyDown(Enum.KeyCode.D) then
		Camera.CFrame = Camera.CFrame * CFrame.fromOrientation(0, -math.rad(PerFrameRotationDelta * RotationAmp), 0);
	end
end);

UIS.InputBegan:Connect(function(io)
	if io.UserInputType ~= Enum.UserInputType.Keyboard then
		return
	end
	--[[
	R move up
	F move down
		                         ^ move forward
		                         W
		         rotate left < A S D > rotate right
		                         v move backward

		                         ^ higher move speed
		                         I
		lower rotation speed < J K L > higher rotation speed
		                         v lower move speed
	]]
	if io.KeyCode == Enum.KeyCode.I then
		MoveAmp = MoveAmp + 1;
	elseif io.KeyCode == Enum.KeyCode.K then
		MoveAmp = math.max(0, MoveAmp - 1);
	end

	if io.KeyCode == Enum.KeyCode.L then
		RotationAmp = RotationAmp + 1;
	elseif io.KeyCode == Enum.KeyCode.J then
		RotationAmp = math.max(0, RotationAmp - 1);
	end
end);
`
